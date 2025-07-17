package sstable

import (
	"encoding/binary"
	"tprgdb/components/blockmanager"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type sstablerecord struct {
	timestamp     uint64 // Timestamp of the fragment
	tombstone     bool   // Tombstone flag
	key           string // Key of the fragment
	compressedKey uint64 // Int version of the key from global compression dictionary
	value         []byte // Values of the fragment
	isCompressed  bool   // Flag to indicate if compression is enabled
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
	// timestamp
	inserted := binary.PutUvarint(variableBuffer, record.timestamp)
	data = append(data, variableBuffer[:inserted]...)
	// tombstone
	data = append(data, serialization.BoolToBytes(record.tombstone)...)
	if record.isCompressed {
		// compressed key
		inserted = binary.PutUvarint(variableBuffer, record.compressedKey)
		data = append(data, variableBuffer[:inserted]...)
	} else {
		// key size
		inserted = binary.PutUvarint(variableBuffer, uint64(len(record.key)))
		data = append(data, variableBuffer[:inserted]...)
		// key
		data = append(data, []byte(record.key)...)
	}
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

	if record.isCompressed {
		// read the compressed key
		compressedKeyBytes, compressedKey, err5 := file.ReadUVariable()
		if err5 != nil {
			return err5
		}
		data = append(data, compressedKeyBytes...)
		record.compressedKey = compressedKey
	} else {
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
	}

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

func (record *sstablerecord) GetTimestamp() uint64 {
	return record.timestamp
}

func (record *sstablerecord) GetTombstone() bool {
	return record.tombstone
}

func (record *sstablerecord) GetKey() string {
	return record.key
}

func (record *sstablerecord) GetCompressedKey() uint64 {
	return record.compressedKey
}

func (record *sstablerecord) GetValue() []byte {
	return record.value
}
