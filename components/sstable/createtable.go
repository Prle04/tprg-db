package sstable

import (
	"os"
	ospath "path"
	"strings"
	"tprgdb/components/blockmanager"
	"tprgdb/components/bloomfilter"
	"tprgdb/components/merkletree"
)

func flushToSSTable(
	iterator ISSTableIterator,
	maximumNumberOfItems uint64,
	folderPath string,
	tableName string,
	summaryDilutionFactor uint8,
	mode string,
	compressionManager ICompressionManager,
	manager *blockmanager.BlockManager,
) (*SSTableReader, error) {

	// Check if the folder exists, create it if it doesn't
	if folderPath != "" {
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			err := os.MkdirAll(folderPath, 0o755)
			if err != nil {
				return nil, err
			}
		}
	}

	if mode != MODE_ONE_FILE && mode != MODE_MULTI_FILE {
		return nil, &IncorrectModeError{mode: mode}
	}
	path := ospath.Join(folderPath, tableName)

	// Create the SSTable Segments
	bloomfilter := bloomfilter.NewBloomFilter(int(maximumNumberOfItems), 0.001)
	index := make(SSTableIndex, maximumNumberOfItems)
	merkleBuilder := merkletree.NewMerkleTreeBuilder()
	// Segment represents a block of data in the SSTable that will be a leaf in the Merkle tree
	buffer := newSSTableBlock(compressionManager.IsCompressionEnabled())
	currentOffset := uint64(0) // current offset in the data file, will be used to write the index
	// Open the data file for writing
	dataFile := manager.OpenBlockedFile(path + "-data.db")
	// We write the SSTable header at the beginning of the file
	dataFile.Seek(SSTableHeaderSize)

	var minKey, maxKey *SSTableIndexItem = nil, nil

	header := &SSTableHeader{
		IndexStartOffset:       0,
		SummaryStartOffset:     0,
		BloomFilterStartOffset: 0,
		MerkleStartOffset:      0,
		NumberOfItems:          0,
		IsCompressionEnabled:   compressionManager.IsCompressionEnabled(),
	}

	// Loop through the items and add them to the segments
	for item, err := iterator.Next(); err == nil; item, err = iterator.Next() {

		var compressedKey uint64 = 0
		key := item.GetKey()
		value := item.GetValue()

		if compressionManager.IsCompressionEnabled() {
			// Get the compressed key
			compressedKey = item.GetCompressedKey()
		}

		// Insert the key into the bloom filter
		bloomfilter.Insert([]byte(key))

		// Add the key to the Index
		index[header.NumberOfItems] = &SSTableIndexItem{
			Key:                  key,
			CompressedKey:        compressedKey,
			Offset:               currentOffset,
			IsCompressionEnabled: compressionManager.IsCompressionEnabled(),
		}

		if minKey == nil || item.GetKey() < minKey.Key {
			minKey = index[header.NumberOfItems]
		}
		if maxKey == nil || item.GetKey() > maxKey.Key {
			maxKey = index[header.NumberOfItems]
		}

		header.NumberOfItems++

		// Will be used to calculate the lenght in bytes
		// that this item will take in the data file
		startingOffset := dataFile.Tell()

		// If the item is deleted we create a special record
		if item.IsDeleted() {
			// While there is no space in the buffer we write the buffer to the file
			// and clear it for the next write
			for {
				ok, err := buffer.delete(dataFile, key, compressedKey, item.GetTimestamp())
				if err != nil {
					return nil, err
				}
				if ok {
					break
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
				remaining, err := buffer.add(dataFile, key, compressedKey, value, item.GetTimestamp(), isContinuance)
				if err != nil {
					return nil, err
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
				isContinuance = len(remaining) < len(value) // If the remaining value is smaller than the original value, it means that we are continuing a previous record
				// If the remaining value has the same size as the original value, it means that we hit the padding threshold
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

	// If the buffer is not empty we need to create a merkle tree leaf from it
	if buffer.size > 0 {
		data := buffer.getData()
		merkleBuilder.AddData(&data)
		buffer.clear()
	}
	
	header.IndexStartOffset = dataFile.Tell()
	// Now that we have completed writting all items to data segments
	// we can write other segments of the SSTable
	// Depending on the mode, they will be written to different files
	if mode == MODE_ONE_FILE {
		// We will write everything to the same file
		currentOffset = 0                // reset the current offset for the index
		summary := make(SSTableIndex, 0) // create a summary of the index

		// Loop through the index and write it to the file
		for i := range header.NumberOfItems {
			item := index[i]
			// Every summaryDilutionFactor-th item will be added to the summary
			if i%uint64(summaryDilutionFactor) == 0 {
				summary = append(summary, &SSTableIndexItem{
					Key:                  item.Key,
					CompressedKey:        item.CompressedKey,
					Offset:               currentOffset,
					IsCompressionEnabled: item.IsCompressionEnabled,
				})
			}
			// Write the index item to the file
			data := item.getData()
			err := dataFile.Write(data)
			if err != nil {
				return nil, err
			}
			// Update the current offset in the data file
			currentOffset += uint64(len(data))
		}
		header.SummaryStartOffset = dataFile.Tell()
		// Write the summary to the file
		// First we write the first and last item of the index so we have range of the whole table
		err := dataFile.Write(minKey.getData()) // Write the first item of the index as a summary
		if err != nil {
			return nil, err
		}
		err = dataFile.Write(maxKey.getData()) // Write the last item of the index as a summary
		if err != nil {
			return nil, err
		}
		// Now we write the summary items
		for _, item := range summary {
			data := item.getData()
			err := dataFile.Write(data)
			if err != nil {
				return nil, err
			}
		}
		// Write the bloom filter to the file
		header.BloomFilterStartOffset = dataFile.Tell()
		err = bloomfilter.SerializeToFile(dataFile)
		if err != nil {
			return nil, err
		}
		// Write the Merkle tree to the file
		header.MerkleStartOffset = dataFile.Tell()
		merkleTree := merkleBuilder.Build()
		if merkleTree == nil {
			return nil, &EmptyMerkleTreeError{}
		}
		err = merkleTree.SerializeToFile(dataFile)
		if err != nil {
			return nil, err
		}
		// Write the SSTable header to the beginning of the file
		dataFile.Seek(0)
		headerData := header.SerializeToBytes()
		err = dataFile.Write(headerData)
		if err != nil {
			return nil, err
		}
		// Close the data file
		err = dataFile.Close()
		if err != nil {
			return nil, err
		}
		// Write the table of contents for this SSTable

		err = os.WriteFile(path+"-toc.txt", []byte(path+"-data.db\n"), 0o644)
		reader := &SSTableReader{
			tableName:          tableName,
			filePath:           path,
			mode:               mode, // "one-file" or "multi-file"
			header:             header,
			blockManager:       manager,
			compressionManager: compressionManager,
			toc:                []string{path + "-data.db\n"},
			filter:             nil,
			summary:            nil,
		}
		return reader, err
	}
	// The mode is "multi-file", we will write the index, bloom filter and Merkle tree to separate files

	// First we write the empty header to the data file
	dataFile.Seek(0)
	headerData := header.SerializeToBytes()
	err := dataFile.Write(headerData)
	if err != nil {
		return nil, err
	}
	// Close the data file
	err = dataFile.Close()
	if err != nil {
		return nil, err
	}

	indexFile := manager.OpenBlockedFile(path + "-index.db")
	summaryFile := manager.OpenBlockedFile(path + "-summary.db")
	bloomFile := manager.OpenBlockedFile(path + "-filter.db")
	merkleFile := manager.OpenBlockedFile(path + "-metadata.db")

	currentOffset = 0 // reset the current offset for the index

	// First we write the first and last item of the index into the summary
	// so we have range of the whole table
	err = summaryFile.Write(minKey.getData()) // Write the first item of the index as a summary
	if err != nil {
		return nil, err
	}
	err = summaryFile.Write(maxKey.getData()) // Write the last item of the index as a summary
	if err != nil {
		return nil, err
	}

	// Loop through the index and write it to the file
	for i, item := range index {
		// Every summaryDilutionFactor-th item will be printed to the summary file
		if i%int(summaryDilutionFactor) == 0 {
			// Write the item to the summary file
			err = summaryFile.Write((&SSTableIndexItem{
				Key:                  item.Key,
				CompressedKey:        item.CompressedKey,
				IsCompressionEnabled: item.IsCompressionEnabled,
				Offset:               currentOffset,
			}).getData())
			if err != nil {
				return nil, err
			}
		}
		// Write the index item to the file
		data := item.getData()
		err := indexFile.Write(data)
		if err != nil {
			return nil, err
		}
		// Update the current offset for the index
		currentOffset += uint64(len(data))
	}
	// Write the bloom filter to the file
	err = bloomfilter.SerializeToFile(bloomFile)
	if err != nil {
		return nil, err
	}
	// Write the Merkle tree to the file
	merkleTree := merkleBuilder.Build()
	err = merkleTree.SerializeToFile(merkleFile)
	if err != nil {
		return nil, err
	}

	// Close all files
	err = indexFile.Close()
	if err != nil {
		return nil, err
	}
	err = summaryFile.Close()
	if err != nil {
		return nil, err
	}
	err = bloomFile.Close()
	if err != nil {
		return nil, err
	}
	err = merkleFile.Close()
	if err != nil {
		return nil, err
	}

	// Write the table of contents for this SSTable

	tocArray := []string{
		path + "-data.db",
		path + "-index.db",
		path + "-summary.db",
		path + "-filter.db",
		path + "-metadata.db",
	}
	tocData := []byte(strings.Join(tocArray, "\n") + "\n")
	err = os.WriteFile(path+"-toc.txt", tocData, 0o644)
	reader := &SSTableReader{
		tableName:          tableName,
		filePath:           path,
		mode:               mode, // "one-file" or "multi-file"
		header:             header,
		blockManager:       manager,
		compressionManager: compressionManager,
		toc:                tocArray,
		filter:             nil,
		summary:            nil,
	}
	return reader, err
}
