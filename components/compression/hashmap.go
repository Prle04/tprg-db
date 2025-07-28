package compression

import (
	"encoding/binary"
	"fmt"
	"tprgdb/components/blockmanager"
)

// COMPRESSION STRUCTURE

type HashMapCompression struct {
	data map[string]uint64
}

func NewHashMapCompression() *HashMapCompression {
	return &HashMapCompression{
		data: make(map[string]uint64),
	}
}

func (h *HashMapCompression) Find(key string) (uint64, error) {
	if value, exists := h.data[key]; exists {
		return value, nil
	}
	return 0, fmt.Errorf("key %s not found in compression dictionary", key)
}
func (h *HashMapCompression) SerializeToBytes() []byte {
	data := make([]byte, 0)
	variableBuffer := make([]byte, 8)
	numberOfEntriesLength := binary.PutUvarint(variableBuffer, uint64(len(h.data)))
	data = append(data, variableBuffer[:numberOfEntriesLength]...)
	for key, value := range h.data {
		keyLenght := binary.PutUvarint(variableBuffer, uint64(len(key)))
		data = append(data, variableBuffer[:keyLenght]...)
		data = append(data, []byte(key)...)
		valueLength := binary.PutUvarint(variableBuffer, value)
		data = append(data, variableBuffer[:valueLength]...)
	}
	return data
}
func (h *HashMapCompression) DeserializeFromBytes(data []byte) {
	numberOfEntries, n := binary.Uvarint(data)
	data = data[n:]
	h.data = make(map[string]uint64, numberOfEntries)
	for range numberOfEntries {
		keyLength, n := binary.Uvarint(data)
		data = data[n:]
		key := string(data[:keyLength])
		data = data[keyLength:]
		value, n := binary.Uvarint(data)
		data = data[n:]
		h.data[key] = value
	}
}
func (h *HashMapCompression) SerializeToFile(file *blockmanager.BlockedFile) error {
	return file.Write(h.SerializeToBytes())
}
func (h *HashMapCompression) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	_, numberOfEntries, err := file.ReadUVariable()
	if err != nil {
		return err
	}
	h.data = make(map[string]uint64, numberOfEntries)
	for i := uint64(0); i < numberOfEntries; i++ {
		_, keyLength, err := file.ReadUVariable()
		if err != nil {
			return err
		}
		keyBytes, err := file.Read(keyLength)
		key := string(keyBytes)
		if err != nil {
			return err
		}
		_, value, err := file.ReadUVariable()
		if err != nil {
			return err
		}
		h.data[key] = value
	}
	return nil
}

func (h *HashMapCompression) Add(key string) uint64 {
	if value, exists := h.data[key]; exists {
		return value
	}
	compressedKey := uint64(len(h.data) + 1) // Simple incremental compression
	h.data[key] = compressedKey
	return compressedKey
}

// DECOMPRESSION STRUCTURE

type HashMapDecompression struct {
	data map[uint64]string
}

func NewHashMapDecompression() *HashMapDecompression {
	return &HashMapDecompression{
		data: make(map[uint64]string),
	}
}
func (h *HashMapDecompression) Find(compressedKey uint64) (string, error) {
	value, exists := h.data[compressedKey]
	if exists {
		return value, nil
	}
	return "", fmt.Errorf("value %d not found in decompression dictionary", compressedKey)
}
func (h *HashMapDecompression) SerializeToBytes() []byte {
	data := make([]byte, 0)
	variableBuffer := make([]byte, 8)
	numberOfEntriesLength := binary.PutUvarint(variableBuffer, uint64(len(h.data)))
	data = append(data, variableBuffer[:numberOfEntriesLength]...)
	for key, value := range h.data {
		keyLength := binary.PutUvarint(variableBuffer, key)
		data = append(data, variableBuffer[:keyLength]...)
		valueLength := binary.PutUvarint(variableBuffer, uint64(len(value)))
		data = append(data, variableBuffer[:valueLength]...)
		data = append(data, []byte(value)...)
	}
	return data
}
func (h *HashMapDecompression) DeserializeFromBytes(data []byte) {
	numberOfEntries, n := binary.Uvarint(data)
	data = data[n:]
	h.data = make(map[uint64]string, numberOfEntries)
	for range numberOfEntries {
		key, n := binary.Uvarint(data)
		data = data[n:]
		valueLength, n := binary.Uvarint(data)
		data = data[n:]
		value := string(data[:valueLength])
		data = data[valueLength:]
		h.data[key] = value
	}
}
func (h *HashMapDecompression) SerializeToFile(file *blockmanager.BlockedFile) error {
	return file.Write(h.SerializeToBytes())
}
func (h *HashMapDecompression) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	_, numberOfEntries, err := file.ReadUVariable()
	if err != nil {
		return err
	}
	h.data = make(map[uint64]string, numberOfEntries)
	for range numberOfEntries {
		_, key, err := file.ReadUVariable()
		if err != nil {
			return err
		}
		_, valueLength, err := file.ReadUVariable()
		if err != nil {
			return err
		}
		valueBytes, err := file.Read(valueLength)
		if err != nil {
			return err
		}
		h.data[key] = string(valueBytes)
	}
	return nil
}

func (h *HashMapDecompression) Add(compressed uint64, decompressed string) {
	h.data[compressed] = decompressed
}
