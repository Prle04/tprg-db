package sstable

import (
	"encoding/binary"
	"tprgdb/components/blockmanager"
)

type SSTableIndexItem struct {
	Key                  string
	CompressedKey        uint64
	Offset               uint64
	IsCompressionEnabled bool
}
type SSTableIndex []*SSTableIndexItem

func (item *SSTableIndexItem) getData() []byte {
	data := make([]byte, 0)
	variableBuffer := make([]byte, 8)
	if item.IsCompressionEnabled {
		// compressed key
		n := binary.PutUvarint(variableBuffer, item.CompressedKey)
		data = append(data, variableBuffer[:n]...)
	} else {
		// key size
		n := binary.PutUvarint(variableBuffer, uint64(len(item.Key)))
		data = append(data, variableBuffer[:n]...)
		// key
		data = append(data, []byte(item.Key)...)
	}
	// offset
	n := binary.PutUvarint(variableBuffer, item.Offset)
	data = append(data, variableBuffer[:n]...)
	return data
}

func (item *SSTableIndexItem) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	var err error
	if item.IsCompressionEnabled {
		// read compressed key
		_, item.CompressedKey, err = file.ReadUVariable()
		if err != nil {
			return err
		}
	} else {
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
	}
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
