package sstable

import "tprgdb/components/blockmanager"

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

func (it *SSIterator) Next() (ISSTableItem, error) {
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
