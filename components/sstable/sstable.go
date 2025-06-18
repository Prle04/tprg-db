package sstable

import (
	"tprgdb/components/blockmanager"
	"tprgdb/components/bloomfilter"
	"tprgdb/components/merkletree"
	"tprgdb/components/serialization"
)

type SSTableItem interface {
	GetKey() string
	GetValue() []byte
	GetTimestamp() uint64
	IsDeleted() bool
}

type SSTableHeader struct {
	IndexStartOffset       uint64
	SummaryStartOffset     uint64
	BloomFilterStartOffset uint64
	MerkleStartOffset      uint64
}

func FlushToSSTable(
	items []SSTableItem,
	path string,
	summaryDilutionFactor uint8,
	mode string,
	manager *blockmanager.BlockManager,
) error {

	if mode != "one-file" && mode != "multi-file" {
		return &IncorrectModeError{mode: mode}
	}

	// Create the SSTable Segments
	bloomfilter := bloomfilter.NewBloomFilter(len(items), 0.001)
	index := make(SSTableIndex, 0, len(items))
	merkleBuilder := merkletree.NewMerkleTreeBuilder()
	// Segment represents a block of data in the SSTable that will be a leaf in the Merkle tree
	buffer := newSSTableBlock()
	currentOffset := uint64(0) // current offset in the data file, will be used to write the index
	// Open the data file for writing
	dataFile := manager.OpenBlockedFile(path + "-data.db")
	// We write the SSTable header at the beginning of the file
	dataFile.Seek(SSTableHeaderSize)

	header := &SSTableHeader{
		IndexStartOffset:       0,
		SummaryStartOffset:     0,
		BloomFilterStartOffset: 0,
		MerkleStartOffset:      0,
	}

	// Loop through the items and add them to the segments
	for i, item := range items {

		// Get the key and value from the item
		key := item.GetKey()
		value := item.GetValue()

		// Insert the key into the bloom filter
		bloomfilter.Insert([]byte(key))

		// Add the key to the Index
		index[i] = SSTableIndexItem{
			Key:    key,
			Offset: currentOffset,
		}

		// Will be used to calculate the lenght in bytes
		// that this item will take in the data file
		startingOffset := dataFile.Tell()

		// If the item is deleted we create a special record
		if item.IsDeleted() {
			// While there is no space in the buffer we write the buffer to the file
			// and clear it for the next write
			for ok, err := buffer.delete(dataFile, key, item.GetTimestamp()); !ok || err != nil; {
				if err != nil {
					return err
				}
				// datablock is full, we need to write it to the file
				data := buffer.getData()
				// Add the data to the Merkle tree buildrer as a leaf node
				merkleBuilder.AddData(&data)
				// Clear the buffer for the next write
				buffer.clear()
			}
		} else {
			// While there is part of the value that does not fit in the buffer
			// we add it to the buffer
			// and write the buffer to the file when it is full
			// and clear it for the next write

			// IsContinuance is used to determine if the record is a continuation of a previous record
			// If it is, we set the record type to middle or last depending on the previous record
			isContinuance := false
			for len(value) > 0 {
				// add the data to the current buffer and get the remaining data for which no space was available
				remaining, err := buffer.add(dataFile, key, value, item.GetTimestamp(), isContinuance)
				if err != nil {
					return err
				}
				// The whole value fitted in the buffer
				if remaining == nil {
					break
				}
				// buffer is full, we need to write it to the file
				data := buffer.getData()
				// Add the data to the Merkle tree builder as a leaf node
				merkleBuilder.AddData(&data)
				// Clear the buffer for the next write
				buffer.clear()
				// We know that the record is not last fragment so we set the type to middle or first depending on the previous record
				isContinuance = true
				// Set the remaining value to the value that did not fit in the buffer
				// and continue the loop to add it to the next buffer
				value = remaining
			}
		}

		// Calculate the lenght in bytes that this item took in the data file
		keyLenght := dataFile.Tell() - startingOffset
		// Update the current offset in the data file
		currentOffset += keyLenght
	}

	// Now that we have completed writting all items to data segments
	// we can write other segments of the SSTable
	// Depending on the mode, they will be written to different files
	if mode == "one-file" {
		// We will write everything to the same file
		header.IndexStartOffset = dataFile.Tell()
		currentOffset = 0                // reset the current offset for the index
		summary := make(SSTableIndex, 0) // create a summary of the index

		// Loop through the index and write it to the file
		for i, item := range index {
			// Every summaryDilutionFactor-th item will be added to the summary
			if i%int(summaryDilutionFactor) == 0 {
				summary = append(summary, SSTableIndexItem{
					Key:    item.Key,
					Offset: currentOffset,
				})
			}
			// Write the index item to the file
			data := item.getData()
			err := dataFile.Write(data)
			if err != nil {
				return err
			}
			// Update the current offset in the data file
			currentOffset += uint64(len(data))
		}
		header.SummaryStartOffset = dataFile.Tell()
		// Write the summary to the file
		// First we write the first and last item of the index so we have range of the whole table
		err := dataFile.Write(index[0].getData()) // Write the first item of the index as a summary
		if err != nil {
			return err
		}
		err = dataFile.Write(index[len(index)-1].getData()) // Write the last item of the index as a summary
		if err != nil {
			return err
		}
		// Now we write the summary items
		for _, item := range summary {
			data := item.getData()
			err := dataFile.Write(data)
			if err != nil {
				return err
			}
		}
		// Write the bloom filter to the file
		header.BloomFilterStartOffset = dataFile.Tell()
		err = bloomfilter.SerializeToFile(dataFile)
		if err != nil {
			return err
		}
		// Write the Merkle tree to the file
		header.MerkleStartOffset = dataFile.Tell()
		merkleTree := merkleBuilder.Build()
		err = merkleTree.SerializeToFile(dataFile)
		if err != nil {
			return err
		}
		// Write the SSTable header to the beginning of the file
		dataFile.Seek(0)
		headerData := header.SerializeToBytes()
		err = dataFile.Write(headerData)
		if err != nil {
			return err
		}
		// Close the data file
		err = dataFile.Close()
		if err != nil {
			return err
		}
		// Write the table of contents for this SSTable
		tocFile := manager.OpenBlockedFile(path + "-toc.txt")
		err = tocFile.Write([]byte(path + "-data.db\n"))
		if err != nil {
			return err
		}
		err = tocFile.Close()
		return err
	}
	// The mode is "multi-file", we will write the index, bloom filter and Merkle tree to separate files

	// First we write the empty header to the data file
	dataFile.Seek(0)
	headerData := header.SerializeToBytes()
	err := dataFile.Write(headerData)
	if err != nil {
		return err
	}
	// Close the data file
	err = dataFile.Close()
	if err != nil {
		return err
	}

	indexFile := manager.OpenBlockedFile(path + "-index.db")
	summaryFile := manager.OpenBlockedFile(path + "-summary.db")
	bloomFile := manager.OpenBlockedFile(path + "-filter.db")
	merkleFile := manager.OpenBlockedFile(path + "-metadata.db")

	currentOffset = 0 // reset the current offset for the index

	// First we write the first and last item of the index into the summary
	// so we have range of the whole table
	err = summaryFile.Write(index[0].getData()) // Write the first item of the index as a summary
	if err != nil {
		return err
	}
	err = summaryFile.Write(index[len(index)-1].getData()) // Write the last item of the index as a summary
	if err != nil {
		return err
	}

	// Loop through the index and write it to the file
	for i, item := range index {
		// Every summaryDilutionFactor-th item will be printed to the summary file
		if i%int(summaryDilutionFactor) == 0 {
			// Write the item to the summary file
			err = summaryFile.Write((SSTableIndexItem{
				Key:    item.Key,
				Offset: currentOffset,
			}).getData())
			if err != nil {
				return err
			}
		}
		// Write the index item to the file
		data := item.getData()
		err := indexFile.Write(data)
		if err != nil {
			return err
		}
		// Update the current offset for the index
		currentOffset += uint64(len(data))
	}
	// Write the bloom filter to the file
	err = bloomfilter.SerializeToFile(bloomFile)
	if err != nil {
		return err
	}
	// Write the Merkle tree to the file
	merkleTree := merkleBuilder.Build()
	err = merkleTree.SerializeToFile(merkleFile)
	if err != nil {
		return err
	}

	// Close all files
	err = indexFile.Close()
	if err != nil {
		return err
	}
	err = summaryFile.Close()
	if err != nil {
		return err
	}
	err = bloomFile.Close()
	if err != nil {
		return err
	}
	err = merkleFile.Close()
	if err != nil {
		return err
	}

	// Write the table of contents for this SSTable
	tocFile := manager.OpenBlockedFile(path + "-toc.txt")
	tocData := []byte(path + "-data.db\n" +
		path + "-index.db\n" +
		path + "-summary.db\n" +
		path + "-filter.db\n" +
		path + "-metadata.db\n")
	err = tocFile.Write(tocData)
	if err != nil {
		return err
	}
	err = tocFile.Close()
	return err

}

// type WALIterator struct {
// 	Record *sstablerecord
// 	Err    error
// }

// func (wal *WriteAheadLog) Iterate() <-chan *WALIterator {
// 	ch := make(chan *WALIterator)
// 	go func() {
// 		defer close(ch)
// 		wal.currentSegment.file.Sync()
// 		entries, err := os.ReadDir(wal.walPath)
// 		if err != nil {
// 			ch <- &WALIterator{Err: err}
// 			return
// 		}
// 		var lastRecord *sstablerecord = nil
// 		for _, v := range entries {
// 			if v.IsDir() {
// 				continue
// 			}
// 			if filepath.Ext(v.Name()) != ".wal" && !strings.HasPrefix(v.Name(), "wal_") {
// 				continue
// 			}
// 			filepath := path.Join(wal.walPath, v.Name())
// 			file := wal.manager.OpenBlockedFile(filepath)
// 			file.Seek(uint64(SegmentNumberSize))
// 			for file.Tell() < wal.segmentSize*wal.blockSize {
// 				record := &sstablerecord{}
// 				err := record.read(file)
// 				if err != nil {
// 					ch <- &WALIterator{Err: err}
// 					return
// 				}
// 				if record.recordType == FragmentTypePadding {
// 					break
// 				}
// 				if record.recordType == FragmentTypeMiddle || record.recordType == FragmentTypeLast {
// 					if lastRecord == nil {
// 						ch <- &WALIterator{Err: &NotFirstFragment{fragmentType: record.recordType}}
// 						return
// 					}
// 					lastRecord.value = append(lastRecord.value, record.value...)
// 					if record.recordType == FragmentTypeLast {
// 						ch <- &WALIterator{Record: lastRecord}
// 						lastRecord = nil
// 					}
// 				} else {
// 					lastRecord = record
// 				}
// 			}
// 		}
// 	}()
// 	return ch
// }

func (header *SSTableHeader) SerializeToBytes() []byte {
	// 8B - IndexStartOffset
	// 8B - SummaryStartOffset
	// 8B - BloomFilterStartOffset
	// 8B - MerkleStartOffset

	result := make([]byte, 0, SSTableHeaderSize)

	// IndexStartOffset
	result = append(result, serialization.UInt64ToBytes(header.IndexStartOffset)...)
	// SummaryStartOffset
	result = append(result, serialization.UInt64ToBytes(header.SummaryStartOffset)...)
	// BloomFilterStartOffset
	result = append(result, serialization.UInt64ToBytes(header.BloomFilterStartOffset)...)
	// MerkleStartOffset
	result = append(result, serialization.UInt64ToBytes(header.MerkleStartOffset)...)

	return result
}
