package sstable

import "tprgdb/components/blockmanager"

type SSTableManager struct {
	folderPath            string
	blockManager          *blockmanager.BlockManager
	compressionManager    ICompressionManager
	summaryDilutionFactor uint8
	mode                  string // "one-file" or "multi-file"
}

func NewSSTableManager(folderPath string, mode string, summaryDilutionFactor uint8, blockManager *blockmanager.BlockManager, compressionManager ICompressionManager) *SSTableManager {
	return &SSTableManager{
		summaryDilutionFactor: summaryDilutionFactor,
		folderPath:            folderPath,
		blockManager:          blockManager,
		compressionManager:    compressionManager,
		mode:                  mode,
	}
}

func (m *SSTableManager) NewSSTableReader(tableName string) *SSTableReader {
	return newSSTableReader(m.folderPath, tableName, m.blockManager, m.compressionManager)
}

func (m *SSTableManager) CreateSSTable(tableName string, dataIterator ISSTableIterator, maxNumberOfElements uint64) (*SSTableReader, error) {
	return flushToSSTable(dataIterator, maxNumberOfElements, m.folderPath, tableName, m.summaryDilutionFactor, m.mode, m.compressionManager, m.blockManager)
}

func (m *SSTableManager) MergeSSTables(newTableName string, readers []*SSTableReader) (*SSTableReader, error) {
	if len(readers) == 0 {
		return nil, nil // No readers to merge
	}
	iterator, err := NewSSTableMergeIterator(readers, m.compressionManager.IsCompressionEnabled())
	if err != nil {
		return nil, err // Error creating merge iterator
	}
	maxNumberOfElements := uint64(0)
	for _, reader := range readers {
		n, err := reader.GetNumberOfItems()
		if err != nil {
			return nil, err // Error getting number of items from reader
		}
		maxNumberOfElements += n
	}
	return m.CreateSSTable(newTableName, iterator, maxNumberOfElements) // Create a new SSTable from the merged iterator
}
