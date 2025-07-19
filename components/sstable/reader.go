package sstable

import (
	"math"
	"os"
	ospath "path"
	"strings"
	"tprgdb/components/blockmanager"
	"tprgdb/components/bloomfilter"
	"tprgdb/components/merkletree"
)

type KeyBoundaries struct {
	MinKey string
	MaxKey string
}

type SSTableReader struct {
	header             *SSTableHeader
	filter             *bloomfilter.BloomFilter
	summary            *SSTableSummary
	blockManager       *blockmanager.BlockManager
	compressionManager ICompressionManager
	toc                []string
	tableName          string
	filePath           string
	mode               string // "one-file" or "multi-file"
	keyBoundaries      *KeyBoundaries
}

func newSSTableReader(
	folderPath string,
	tableName string,
	blockManager *blockmanager.BlockManager,
	compressionManager ICompressionManager,
) *SSTableReader {
	path := ospath.Join(folderPath, tableName)
	table := &SSTableReader{
		tableName:          tableName,
		filePath:           path,
		blockManager:       blockManager,
		compressionManager: compressionManager,
		header:             nil,
		filter:             nil,
		summary:            nil,
		toc:                nil,
	}
	toc, err := table.loadTOC()
	if err == nil {
		table.toc = toc
		if len(toc) == 1 {
			table.mode = MODE_ONE_FILE
		} else {
			table.mode = MODE_MULTI_FILE
		}
	}
	header, err := table.loadHeader()
	if err == nil {
		table.header = header
		if header.IndexStartOffset == 0 && header.BloomFilterStartOffset == 0 && header.SummaryStartOffset == 0 && header.MerkleStartOffset == 0 {
			// If all offsets are zero, it indicates that segments are in separate files
			table.mode = MODE_MULTI_FILE
		}
	} else {
		table.mode = MODE_ONE_FILE // Default to one-file mode if header cannot be loaded
	}
	return table
}

func (s *SSTableReader) Get(key string) (*SSItem, error) {

	defer func(ss *SSTableReader) {
		ss.filter = nil  // Reset the filter to avoid unnecessary reads
		ss.summary = nil // Reset the summary to avoid unnecessary reads
	}(s)

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
	if s.mode == MODE_ONE_FILE {
		// mode is "one-file"
		indexFile = s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
		offset = current.Offset + header.IndexStartOffset // Total offset in the data file
	} else {
		// mode is "multi-file"
		indexFile = s.blockManager.OpenBlockedFile(s.filePath + "-index.db")
		offset = current.Offset
	}
	// Seek to the offset in the index file
	indexFile.Seek(offset)
	item := &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
	for {
		err = item.DeserializeFromFile(indexFile)
		if err != nil {
			if err == blockmanager.EOF {
				return nil, nil // Key not found in the index
			}
			return nil, err
		}
		if s.IsCompressed() {
			// Decompress the key if compression is enabled
			decompressedKey, err := s.compressionManager.GetDecompressed(item.CompressedKey)
			if err != nil {
				return nil, err
			}
			item.Key = decompressedKey
		}
		if item.Key == key {
			break // Found the item with the key
		}
		if item.Key > key {
			return nil, nil // Key not found in the index
		}
	}

	valueFile := s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
	valueFile.Seek(item.Offset + SSTableHeaderSize) // Seek to the value offset in the data file
	result := &SSItem{
		Key:           item.Key,
		CompressedKey: item.CompressedKey,
		Tombstone:     false,
		Timestamp:     0,
		Value:         make([]byte, 0),
	}
	record := &sstablerecord{isCompressed: header.IsCompressionEnabled}
	err = record.read(valueFile)
	if err != nil {
		return nil, err
	}
	result.Value = record.value
	result.Tombstone = record.tombstone
	result.Timestamp = record.timestamp
	if s.IsCompressed() {
		// Decompress the key if compression is enabled
		decompressedKey, err := s.compressionManager.GetDecompressed(result.CompressedKey)
		if err != nil {
			return nil, err
		}
		result.Key = decompressedKey
	}
	return result, nil
}

func (s *SSTableReader) GetNumberOfItems() (uint64, error) {
	header, err := s.getHeader()
	if err != nil {
		return 0, err
	}
	return header.NumberOfItems, nil
}

func (s *SSTableReader) IsCompressed() bool {
	header, err := s.getHeader()
	if err != nil {
		return false // If header cannot be loaded, assume not compressed
	}
	return header.IsCompressionEnabled
}

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
	if s.mode == MODE_ONE_FILE {
		// mode is "one-file"
		header := &SSTableHeader{}
		dataFile := s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
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
		bloomFile := s.blockManager.OpenBlockedFile(s.filePath + "-filter.db")
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
	dataFile := s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
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

	var summaryFile *blockmanager.BlockedFile
	if s.mode == MODE_ONE_FILE {
		summaryFile = s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
		summaryFile.Seek(header.SummaryStartOffset)
	} else {
		summaryFile = s.blockManager.OpenBlockedFile(s.filePath + "-summary.db")
	}
	s.summary.minKey = &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
	err = s.summary.minKey.DeserializeFromFile(summaryFile)
	if err != nil {
		return nil, err
	}
	s.summary.maxKey = &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
	err = s.summary.maxKey.DeserializeFromFile(summaryFile)
	if err != nil {
		return nil, err
	}
	for summaryFile.Tell() < header.BloomFilterStartOffset {
		item := &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
		err = item.DeserializeFromFile(summaryFile)
		if err != nil {
			if err == blockmanager.EOF {
				break
			}
			return nil, err
		}
		if header.IsCompressionEnabled {
			// Decompress the key if compression is enabled
			decompressedKey, err := s.compressionManager.GetDecompressed(item.CompressedKey)
			if err != nil {
				return nil, err
			}
			item.Key = decompressedKey
		}
		s.summary.index = append(s.summary.index, item)
	}
	summaryFile.Close()

	if header.IsCompressionEnabled {
		// Decompress Min and Max keys if compression is enabled
		minKey, err := s.compressionManager.GetDecompressed(s.summary.minKey.CompressedKey)
		if err != nil {
			return nil, err
		}
		s.summary.minKey.Key = minKey
		maxKey, err := s.compressionManager.GetDecompressed(s.summary.maxKey.CompressedKey)
		if err != nil {
			return nil, err
		}
		s.summary.maxKey.Key = maxKey
	}
	return s.summary, nil
}

func (s *SSTableReader) GetIterator() ISSTableIterator {
	return NewSSIterator(s, s.compressionManager)
}

func (s *SSTableReader) GetTableName() string {
	return s.tableName
}

func (s *SSTableReader) Delete() error {
	toDelete := []string{
		s.filePath + "-data.db",
	}
	if s.mode == MODE_MULTI_FILE {
		toDelete = append(toDelete, s.filePath+"-index.db")
		toDelete = append(toDelete, s.filePath+"-filter.db")
		toDelete = append(toDelete, s.filePath+"-summary.db")
		toDelete = append(toDelete, s.filePath+"-metadata.db")
	}
	var resultError error
	for _, filepath := range toDelete {
		err := s.blockManager.OpenBlockedFile(filepath).Delete()
		if err != nil {
			resultError = err
		}
	}
	err := os.Remove(s.filePath + "-toc.txt")
	if err != nil && !os.IsNotExist(err) {
		resultError = err
	}
	return resultError
}

func (s *SSTableReader) getDataFile() *blockmanager.BlockedFile {
	file := s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
	file.Seek(SSTableHeaderSize) // Skip the header
	return file
}

func (s *SSTableReader) GetFilePath() string {
	return s.filePath
}

func (s *SSTableReader) Validate() (bool, error) {
	header, err := s.getHeader()
	if err != nil {
		return false, err
	}
	var merkleFile *blockmanager.BlockedFile
	if s.mode == MODE_ONE_FILE {
		// mode is "one-file"
		merkleFile = s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
		merkleFile.Seek(header.MerkleStartOffset)
	} else {
		// mode is "multi-file"
		merkleFile = s.blockManager.OpenBlockedFile(s.filePath + "-metadata.db")
	}
	merkleTree := &merkletree.MerkleTree{}
	err = merkleTree.DeserializeFromFile(merkleFile)
	if err != nil {
		merkleFile.Close()
	}
	merkleFile.Close()
	dataFile := s.getDataFile()
	builder := merkletree.NewMerkleTreeBuilder()
	for dataFile.Tell() < header.IndexStartOffset {
		toRead := uint64(math.Min(float64(SegmentBlockSize), float64(header.IndexStartOffset-dataFile.Tell())))
		buffer, err := dataFile.Read(toRead)
		if err != nil {
			return false, err
		}
		builder.AddData(&buffer)
	}

	dataFile.Close()
	newTree := builder.Build()
	return newTree.GetRoot() == merkleTree.GetRoot(), nil // Compare the roots of the trees
}

func (s *SSTableReader) GetKeyBoundaries() (*KeyBoundaries, error) {
	if s.keyBoundaries != nil {
		return s.keyBoundaries, nil
	}
	s.summary = &SSTableSummary{}
	header, err := s.getHeader()
	if err != nil {
		return nil, err
	}
	var file *blockmanager.BlockedFile
	if s.mode == MODE_ONE_FILE {
		file = s.blockManager.OpenBlockedFile(s.filePath + "-data.db")
		file.Seek(header.SummaryStartOffset)
	} else {
		file = s.blockManager.OpenBlockedFile(s.filePath + "-summary.db")
	}

	s.summary.minKey = &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
	err = s.summary.minKey.DeserializeFromFile(file)
	if err != nil {
		return nil, err
	}
	s.summary.maxKey = &SSTableIndexItem{IsCompressionEnabled: header.IsCompressionEnabled}
	err = s.summary.maxKey.DeserializeFromFile(file)
	if err != nil {
		return nil, err
	}
	file.Close()
	if header.IsCompressionEnabled {
		// Decompress Min and Max keys if compression is enabled
		minKey, err := s.compressionManager.GetDecompressed(s.summary.minKey.CompressedKey)
		if err != nil {
			return nil, err
		}
		s.summary.minKey.Key = minKey
		maxKey, err := s.compressionManager.GetDecompressed(s.summary.maxKey.CompressedKey)
		if err != nil {
			return nil, err
		}
		s.summary.maxKey.Key = maxKey
	}
	s.keyBoundaries = &KeyBoundaries{
		MinKey: s.summary.minKey.Key,
		MaxKey: s.summary.maxKey.Key,
	}
	// Return the key boundaries
	return s.keyBoundaries, nil
}

func (s *SSTableReader) AreKeysOverlaping(other *SSTableReader) (bool, error) {
	currentBoundaries, err := s.GetKeyBoundaries()
	if err != nil {
		return false, err
	}
	otherBoundaries, err := other.GetKeyBoundaries()
	if err != nil {
		return false, err
	}
	return currentBoundaries.MinKey <= otherBoundaries.MaxKey && otherBoundaries.MinKey <= currentBoundaries.MaxKey, nil
}
