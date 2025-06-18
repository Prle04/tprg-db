package sstable

import (
	"encoding/binary"
	"tprgdb/components/blockmanager"
)

type SSTableIndexItem struct {
	Key    string
	Offset uint64
}
type SSTableIndex []*SSTableIndexItem

func (item *SSTableIndexItem) getData() []byte {
	data := make([]byte, 0)
	variableBuffer := make([]byte, 8)
	// key size
	n := binary.PutUvarint(variableBuffer, uint64(len(item.Key)))
	data = append(data, variableBuffer[:n]...)
	// key
	data = append(data, []byte(item.Key)...)
	// offset
	n = binary.PutUvarint(variableBuffer, item.Offset)
	data = append(data, variableBuffer[:n]...)
	return data
}

func (item *SSTableIndexItem) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	_, keyLenght, err := file.ReadUVariable()
	if err != nil {
		return err
	}
	if keyLenght == 0 {
		return blockmanager.EOF
	}
	keyBytes, err := file.Read(keyLenght)
	if err != nil {
		return err
	}
	item.Key = string(keyBytes)
	_, item.Offset, err = file.ReadUVariable()
	if err != nil {
		return err
	}
	return nil
}

type SSTableSummary struct {
	index  SSTableIndex
	minKey *SSTableIndexItem
	maxKey *SSTableIndexItem
}
