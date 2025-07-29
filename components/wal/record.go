package wal

import (
	"fmt"
	"tprgdb/components/blockmanager"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type walrecord struct {
	recordType uint8  // Type of the fragment
	timestamp  uint64 // Timestamp of the fragment
	tombstone  bool   // Tombstone flag
	key        string // Key of the fragment
	value      []byte // Values of the fragment
}

func (record *walrecord) size() uint64 {
	var size uint64 = uint64(CRCSize) + uint64(TypeSize) + uint64(TimestampSize) + uint64(TombstoneSize) + uint64(KeySize) + uint64(len(record.key))
	if !record.tombstone {
		size += uint64(ValueSize) + uint64(len(record.value))
	}
	return size
}

func (record *walrecord) write(file *blockmanager.BlockedFile) error {
	hash := hash.HashWithSeed{Seed: HashSeed}
	data := make([]byte, 0)
	// type
	data = append(data, serialization.UInt8ToBytes(record.recordType)...)
	// timestamp
	data = append(data, serialization.UInt64ToBytes(record.timestamp)...)
	// tombstone
	data = append(data, serialization.BoolToBytes(record.tombstone)...)
	// key size
	data = append(data, serialization.UInt64ToBytes(uint64(len(record.key)))...)
	// key
	data = append(data, []byte(record.key)...)
	if !record.tombstone {
		// value size
		data = append(data, serialization.UInt64ToBytes(uint64(len(record.value)))...)
		// value
		data = append(data, record.value...)
	}
	// crc
	crc := serialization.UInt64ToBytes(hash.Hash(data))
	data = append(crc, data...)
	err := file.Write(data)
	return err
}

func (record *walrecord) read(file *blockmanager.BlockedFile) error {
	hash := hash.HashWithSeed{Seed: HashSeed}
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
	timestampBytes, err3 := file.Read(uint64(TimestampSize))
	if err3 != nil {
		return err3
	}
	data = append(data, timestampBytes...)
	record.timestamp = serialization.BytesToUInt64(timestampBytes)

	// read the tombstone flag
	tombstoneByte, err4 := file.Read(uint64(TombstoneSize))
	if err4 != nil {
		return err4
	}
	data = append(data, tombstoneByte...)
	record.tombstone = serialization.BytesToBool(tombstoneByte)

	// read the key
	keyLenghtBytes, err5 := file.Read(uint64(KeySize))
	if err5 != nil {
		return err5
	}
	data = append(data, keyLenghtBytes...)
	keyLength := serialization.BytesToUInt64(keyLenghtBytes)
	keyBytes, err6 := file.Read(keyLength)
	if err6 != nil {
		return err6
	}
	data = append(data, keyBytes...)
	record.key = serialization.BytesToString(keyBytes)

	// if the fragment is deleted there is noting else to read
	if !record.tombstone {
		// read the value
		valueLengthBytes, err7 := file.Read(uint64(ValueSize))
		if err7 != nil {
			return err7
		}
		data = append(data, valueLengthBytes...)
		valueLength := serialization.BytesToUInt64(valueLengthBytes)
		valueBytes, err8 := file.Read(valueLength)
		if err8 != nil {
			return err8
		}
		data = append(data, valueBytes...)
		record.value = valueBytes
	}
	if crc != hash.Hash(data) {
		return &CRCError{}
	}
	return nil
}

func (record *walrecord) GetType() uint8 {
	return record.recordType
}

func (record *walrecord) GetTimestamp() uint64 {
	return record.timestamp
}

func (record *walrecord) GetTombstone() bool {
	return record.tombstone
}

func (record *walrecord) GetKey() string {
	return record.key
}

func (record *walrecord) GetValue() []byte {
	return record.value
}

func (record *walrecord) Print() {
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
