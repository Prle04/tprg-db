package sstable

import (
	"encoding/binary"
	"fmt"
	"tprgdb/components/blockmanager"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type sstablerecord struct {
	recordType uint8  // Type of the fragment
	timestamp  uint64 // Timestamp of the fragment
	tombstone  bool   // Tombstone flag
	key        string // Key of the fragment
	value      []byte // Values of the fragment
}

func (record *sstablerecord) size() uint64 {
	return uint64(len(record.getData()))
}

func (record *sstablerecord) write(file *blockmanager.BlockedFile) error {
	data := record.getData()
	// Write the data to the file
	err := file.Write(data)
	return err
}

func (record *sstablerecord) getData() []byte {
	data := make([]byte, 0)
	variableBuffer := make([]byte, 8)
	// type
	data = append(data, serialization.UInt8ToBytes(record.recordType)...)
	// timestamp
	inserted := binary.PutUvarint(variableBuffer, record.timestamp)
	data = append(data, variableBuffer[:inserted]...)
	// tombstone
	data = append(data, serialization.BoolToBytes(record.tombstone)...)
	// key size
	inserted = binary.PutUvarint(variableBuffer, uint64(len(record.key)))
	data = append(data, variableBuffer[:inserted]...)
	// key
	data = append(data, []byte(record.key)...)
	if !record.tombstone {
		// value size
		inserted = binary.PutUvarint(variableBuffer, uint64(len(record.value)))
		data = append(data, variableBuffer[:inserted]...)
		// value
		data = append(data, record.value...)
	}
	// crc
	crc := serialization.UInt64ToBytes(hash.GetMD5Hash(data))
	data = append(crc, data...)
	return data
}

func (record *sstablerecord) read(file *blockmanager.BlockedFile) error {
	// read the crc
	crcBytes, err1 := file.Read(uint64(CRCSize))
	if err1 != nil {
		return err1
	}
	crc := serialization.BytesToUInt64(crcBytes)
	data := make([]byte, 0)
	// read the type
	typeByte, err2 := file.Read(uint64(TypeSize))
	if err2 != nil {
		return err2
	}
	data = append(data, typeByte...)
	record.recordType = serialization.BytesToUInt8(typeByte)
	// if the fragment is padding, we reached the end of the segment
	if record.recordType == FragmentTypePadding {
		return nil
	}

	// read the timestamp
	timestampBytes, timestampValue, err3 := file.ReadUVariable()
	if err3 != nil {
		return err3
	}
	data = append(data, timestampBytes...)
	record.timestamp = timestampValue

	// read the tombstone flag
	tombstoneByte, err4 := file.Read(uint64(TombstoneSize))
	if err4 != nil {
		return err4
	}
	data = append(data, tombstoneByte...)
	record.tombstone = serialization.BytesToBool(tombstoneByte)

	// read the key
	keyLenghtBytes, keyLength, err5 := file.ReadUVariable()
	if err5 != nil {
		return err5
	}
	data = append(data, keyLenghtBytes...)
	keyBytes, err6 := file.Read(keyLength)
	if err6 != nil {
		return err6
	}
	data = append(data, keyBytes...)
	record.key = serialization.BytesToString(keyBytes)

	// if the fragment is deleted there is noting else to read
	if !record.tombstone {
		// read the value
		valueLengthBytes, valueLength, err7 := file.ReadUVariable()
		if err7 != nil {
			return err7
		}
		data = append(data, valueLengthBytes...)
		valueBytes, err8 := file.Read(valueLength)
		if err8 != nil {
			return err8
		}
		data = append(data, valueBytes...)
		record.value = valueBytes
	}
	if crc != hash.GetMD5Hash(data) {
		return &CRCError{}
	}
	return nil
}

func (record *sstablerecord) GetType() uint8 {
	return record.recordType
}

func (record *sstablerecord) GetTimestamp() uint64 {
	return record.timestamp
}

func (record *sstablerecord) GetTombstone() bool {
	return record.tombstone
}

func (record *sstablerecord) GetKey() string {
	return record.key
}

func (record *sstablerecord) GetValue() []byte {
	return record.value
}

func (record *sstablerecord) Print() {
	switch record.recordType {
	case FragmentTypeFull:
		println("Type: Full")
	case FragmentTypeFirst:
		println("Type: First")
	case FragmentTypeMiddle:
		println("Type: Middle")
	case FragmentTypeLast:
		println("Type: Last")
	case FragmentTypePadding:
		println("Type: Padding")
	}
	println("Timestamp: ", record.timestamp)
	fmt.Println("Key: ", record.key)
	println("Tombstone: ", record.tombstone)
	if !record.tombstone {
		fmt.Println("Value: ", string(record.value))
	}
	fmt.Println("====================================")
}
