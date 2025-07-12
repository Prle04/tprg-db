package sstable

import (
	"os"
	ospath "path"
	"strings"
	"tprgdb/components/blockmanager"
	"tprgdb/components/bloomfilter"
)

type SSTableReader struct {
	header   *SSTableHeader
	filter   *bloomfilter.BloomFilter
	summary  *SSTableSummary
	manager  *blockmanager.BlockManager
	index    *SSTableIndex
	toc      []string
	filePath string
	mode     string // "one-file" or "multi-file"
}

type SSItem struct {
	Key       string
	Value     []byte
	Tombstone bool
	Timestamp uint64
}

type SSIterator struct {
	file               *blockmanager.BlockedFile
	numberOfReadItems  uint64
	totalNumberOfItems uint64
}

func NewSSTableReader(
	folderPath string,
	tableName string,
	manager *blockmanager.BlockManager,
) *SSTableReader {
	path := ospath.Join(folderPath, tableName)
	table := &SSTableReader{
		filePath: path,
		manager:  manager,
		index:    nil,
		header:   nil,
		filter:   nil,
		summary:  nil,
		toc:      nil,
	}
	toc, err := table.loadTOC()
	if err == nil {
		table.toc = toc
		if len(toc) == 1 {
			table.mode = "one-file"
		} else {
			table.mode = "multi-file"
		}
	}
	header, err := table.loadHeader()
	if err == nil {
		table.header = header
		if header.IndexStartOffset == 0 && header.BloomFilterStartOffset == 0 && header.SummaryStartOffset == 0 && header.MerkleStartOffset == 0 {
			// If all offsets are zero, it indicates that segments are in separate files
			table.mode = "multi-file"
		}
	} else {
		table.mode = "one-file"
	}
	return table
}

func (s *SSTableReader) Get(key string) (*SSItem, error) {
	filter, err := s.getBloomFilter()
	if err != nil {
		return nil, err
	}
	if filter.IsNotIn([]byte(key)) {
		return nil, nil // Key is not in the Bloom filter, hence not in the SSTable
	}
	summary, err := s.getSummary()
	if err != nil {
		return nil, err
	}
	if key < summary.minKey.Key || key > summary.maxKey.Key {
		return nil, nil // Key is out of range of the SSTable
	}
	header, err := s.getHeader()
	if err != nil {
		return nil, err
	}
	current := summary.index[0]
	for _, item := range summary.index {
		if key >= item.Key {
			break
		}
		current = item
	}
	// File that will be used to read the index
	var indexFile *blockmanager.BlockedFile
	// Offset in the index file where the item starts
	var offset uint64
	if s.mode == "one-file" {
		// mode is "one-file"
		indexFile = s.manager.OpenBlockedFile(s.filePath + "-data.db")
		offset = current.Offset + header.IndexStartOffset // Total offset in the data file
	} else {
		// mode is "multi-file"
		indexFile = s.manager.OpenBlockedFile(s.filePath + "-index.db")
		offset = current.Offset
	}
	// Seek to the offset in the index file
	indexFile.Seek(offset)
	item := &SSTableIndexItem{}
	for {
		err = item.DeserializeFromFile(indexFile)
		if err != nil {
			if err == blockmanager.EOF {
				return nil, nil // Key not found in the index
			}
			return nil, err
		}
		if item.Key == key {
			break
		}
		if item.Key > key {
			return nil, nil // Key not found in the index
		}
	}

	valueFile := s.manager.OpenBlockedFile(s.filePath + "-data.db")
	valueFile.Seek(item.Offset + SSTableHeaderSize) // Seek to the value offset in the data file
	result := &SSItem{
		Key:       item.Key,
		Tombstone: false,
		Timestamp: 0,
		Value:     make([]byte, 0),
	}
	record := &sstablerecord{}
	err = record.read(valueFile)
	if err != nil {
		return nil, err
	}
	result.Value = record.value
	result.Tombstone = record.tombstone
	result.Timestamp = record.timestamp
	switch record.recordType {
	case FragmentTypePadding:
		// Padding record, skip it
		return nil, nil
	case FragmentTypeFull:
		// Full record, return it
		return result, nil
	case FragmentTypeFirst:
		for record.recordType != FragmentTypeLast {
			err = record.read(valueFile)
			if err != nil {
				return nil, err
			}
			result.Value = append(result.Value, record.value...)
		}
		return result, nil
	default:
		return nil, &NotFirstFragment{fragmentType: record.recordType}
	}
}

func (s *SSTableReader) GetNumberOfItems() (uint64, error) {
	header, err := s.getHeader()
	if err != nil {
		return 0, err
	}
	return header.NumberOfItems, nil
}

// func (s *SSTableReader) getTOC() ([]string, error) {
// 	if s.toc != nil {
// 		return s.toc, nil
// 	}
// 	toc, err := s.loadTOC()
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.toc = toc
// 	if len(toc) == 1 {
// 		s.mode = "one-file"
// 	} else {
// 		s.mode = "multi-file"
// 	}
// 	return toc, nil
// }

func (s *SSTableReader) getHeader() (*SSTableHeader, error) {
	if s.header != nil {
		return s.header, nil
	}
	header, err := s.loadHeader()
	if err != nil {
		return nil, err
	}
	s.header = header
	return header, nil
}

func (s *SSTableReader) getBloomFilter() (*bloomfilter.BloomFilter, error) {
	if s.filter != nil {
		return s.filter, nil
	}
	filter, err := s.loadBloomFilter()
	if err != nil {
		return nil, err
	}
	s.filter = filter
	return filter, nil
}

func (s *SSTableReader) getSummary() (*SSTableSummary, error) {
	if s.summary != nil {
		return s.summary, nil
	}
	summary, err := s.loadSummary()
	if err != nil {
		return nil, err
	}
	s.summary = summary
	return summary, nil
}

func (s *SSTableReader) loadBloomFilter() (*bloomfilter.BloomFilter, error) {

	filter := bloomfilter.NewEmptyBloomFilter()
	if s.mode == "one-file" {
		// mode is "one-file"
		header := &SSTableHeader{}
		dataFile := s.manager.OpenBlockedFile(s.filePath + "-data.db")
		headerData, err := dataFile.Read(SSTableHeaderSize)
		if err != nil {
			return nil, err
		}
		header.DeserializeFromBytes(headerData)
		dataFile.Seek(header.BloomFilterStartOffset)
		err = filter.DeserializeFromFile(dataFile)
		if err != nil {
			return nil, err
		}
	} else {
		// mode is "multi-file"
		bloomFile := s.manager.OpenBlockedFile(s.filePath + "-filter.db")
		err := filter.DeserializeFromFile(bloomFile)
		if err != nil {
			return nil, err
		}
	}
	return filter, nil
}

func (s *SSTableReader) loadTOC() ([]string, error) {

	data, err := os.ReadFile(s.filePath + "-toc.txt")
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
	// Binary mode is not used here, we read the file as text
	// tocFile := manager.OpenBlockedFile(filePath + "-toc.txt")
	// lines := make([]string, 0)
	// line := make([]byte, 0, 1024)

	// for {
	// 	b, err := tocFile.Read(1)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if b[0] == 0 {
	// 		break // EOF
	// 	}
	// 	char := string(b)
	// 	if char == "\n" {
	// 		if len(line) == 0 {
	// 			continue
	// 		}
	// 		lines = append(lines, string(line))
	// 		line = make([]byte, 0, 1024)
	// 	} else {
	// 		line = append(line, b...)
	// 	}
	// }
	// return lines, nil
}

func (s *SSTableReader) loadHeader() (*SSTableHeader, error) {
	dataFile := s.manager.OpenBlockedFile(s.filePath + "-data.db")
	headerData, err := dataFile.Read(SSTableHeaderSize)
	if err != nil {
		return nil, err
	}
	s.header = &SSTableHeader{}
	s.header.DeserializeFromBytes(headerData)
	return s.header, nil
}

func (s *SSTableReader) loadSummary() (*SSTableSummary, error) {
	s.summary = &SSTableSummary{}
	header, err := s.getHeader()
	if err != nil {
		return nil, err
	}

	if s.mode == "one-file" {
		dataFile := s.manager.OpenBlockedFile(s.filePath + "-data.db")
		dataFile.Seek(header.SummaryStartOffset)
		s.summary.minKey = &SSTableIndexItem{}
		err = s.summary.minKey.DeserializeFromFile(dataFile)
		if err != nil {
			return nil, err
		}
		s.summary.maxKey = &SSTableIndexItem{}
		err = s.summary.maxKey.DeserializeFromFile(dataFile)
		if err != nil {
			return nil, err
		}
		for dataFile.Tell() < header.BloomFilterStartOffset {
			item := &SSTableIndexItem{}
			err = item.DeserializeFromFile(dataFile)
			if err != nil {
				return nil, err
			}
			s.summary.index = append(s.summary.index, item)
		}
		dataFile.Close()
	} else {
		summaryFile := s.manager.OpenBlockedFile(s.filePath + "-summary.db")
		s.summary.minKey = &SSTableIndexItem{}
		err = s.summary.minKey.DeserializeFromFile(summaryFile)
		if err != nil {
			return nil, err
		}
		s.summary.maxKey = &SSTableIndexItem{}
		err = s.summary.maxKey.DeserializeFromFile(summaryFile)
		if err != nil {
			return nil, err
		}
		for {
			item := &SSTableIndexItem{}
			err = item.DeserializeFromFile(summaryFile)
			if err != nil {
				if err == blockmanager.EOF {
					break
				}
				return nil, err
			}
			s.summary.index = append(s.summary.index, item)
		}
		summaryFile.Close()
	}
	return s.summary, nil
}

func (s *SSTableReader) GetIterator() (*SSIterator, error) {
	file := s.manager.OpenBlockedFile(s.filePath + "-data.db")
	header, err := s.getHeader()
	if err != nil {
		return nil, err
	}
	file.Seek(SSTableHeaderSize) // Skip the header
	return &SSIterator{
		file:               file,
		numberOfReadItems:  0,
		totalNumberOfItems: header.NumberOfItems,
	}, nil
}

func (it *SSIterator) Next() (*SSItem, error) {
	if it.numberOfReadItems >= it.totalNumberOfItems {
		return nil, blockmanager.EOF // No more items to read
	}
	record := &sstablerecord{}
	err := record.read(it.file)
	if err != nil {
		return nil, err
	}
	item := &SSItem{
		Key:       record.key,
		Value:     record.value,
		Tombstone: record.tombstone,
		Timestamp: record.timestamp,
	}
	it.numberOfReadItems++
	return item, nil
}

func (it *SSIterator) Reset() {
	it.file.Seek(SSTableHeaderSize) // Reset to the start of the data file after the header
	it.numberOfReadItems = 0
}

func (item *SSItem) GetKey() string {
	return item.Key
}

func (item *SSItem) GetValue() []byte {
	return item.Value
}

func (item *SSItem) IsDeleted() bool {
	return item.Tombstone
}

func (item *SSItem) GetTimestamp() uint64 {
	return item.Timestamp
}
