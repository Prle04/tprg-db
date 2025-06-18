package sstable

import "encoding/binary"

type SSTableIndexItem struct {
	Key    string
	Offset uint64
}
type SSTableIndex []SSTableIndexItem

func (item SSTableIndexItem) getData() []byte {
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
