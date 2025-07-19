package sstable

import (
	"tprgdb/components/blockmanager"
)

type ISSTableItem interface {
	GetKey() string
	GetCompressedKey() uint64
	GetValue() []byte
	GetTimestamp() uint64
	IsDeleted() bool
}

type ISSTableIterator interface {
	Next() (ISSTableItem, error)
}

// SSItem represents a single item in the SSTable.

type SSItem struct {
	Key           string
	CompressedKey uint64
	Value         []byte
	Tombstone     bool
	Timestamp     uint64
}

func (item SSItem) GetKey() string {
	return item.Key
}
func (item SSItem) GetCompressedKey() uint64 {
	return item.CompressedKey
}
func (item SSItem) GetValue() []byte {
	return item.Value
}
func (item SSItem) IsDeleted() bool {
	return item.Tombstone
}
func (item SSItem) GetTimestamp() uint64 {
	return item.Timestamp
}

// Iterator for reading a file without changing the order of items
type SSIterator struct {
	reader             *SSTableReader
	file               *blockmanager.BlockedFile
	compressionManager ICompressionManager
	numberOfReadItems  uint64
}

func NewSSIterator(reader *SSTableReader, compressionManager ICompressionManager) *SSIterator {
	return &SSIterator{
		reader:             reader,
		file:               reader.getDataFile(),
		compressionManager: compressionManager,
		numberOfReadItems:  0,
	}
}

func (it *SSIterator) Next() (ISSTableItem, error) {
	numberOfItems, err := it.reader.GetNumberOfItems()
	if err != nil {
		return nil, err
	}
	if it.numberOfReadItems >= numberOfItems {
		return nil, blockmanager.EOF // No more items to read
	}
	record := &sstablerecord{isCompressed: it.reader.IsCompressed()}
	err = record.read(it.file)
	if err != nil {
		return nil, err
	}
	item := &SSItem{
		Key:           record.key,
		CompressedKey: record.compressedKey,
		Value:         record.value,
		Tombstone:     record.tombstone,
		Timestamp:     record.timestamp,
	}

	if it.reader.IsCompressed() {
		decompressedKey, err := it.compressionManager.GetDecompressed(item.CompressedKey)
		if err != nil {
			return nil, err
		}
		item.Key = decompressedKey
	} else {
		compressedKey, err := it.compressionManager.GetCompressed(item.Key)
		if err != nil {
			return nil, err
		}
		item.CompressedKey = compressedKey
	}
	it.numberOfReadItems++
	return item, nil
}

type SSListIterator struct {
	items              []ISSTableItem
	compressionManager ICompressionManager
	index              int
}

func NewSSListIterator(items []ISSTableItem, compressionManager ICompressionManager) *SSListIterator {
	return &SSListIterator{
		items:              items,
		index:              0,
		compressionManager: compressionManager,
	}
}

func (it *SSListIterator) Next() (ISSTableItem, error) {
	if it.index >= len(it.items) {
		return nil, blockmanager.EOF // No more items to read
	}
	item := it.items[it.index]
	ssitem := &SSItem{
		Key:           item.GetKey(),
		CompressedKey: item.GetCompressedKey(),
		Value:         item.GetValue(),
		Tombstone:     item.IsDeleted(),
		Timestamp:     item.GetTimestamp(),
	}
	it.index++
	if it.compressionManager.IsCompressionEnabled() {
		compressedKey, err := it.compressionManager.GetCompressed(item.GetKey())
		if err != nil {
			return nil, err
		}
		ssitem.CompressedKey = compressedKey
	}
	return ssitem, nil
}
