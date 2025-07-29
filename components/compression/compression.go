package compression

import (
	"tprgdb/components/blockmanager"
	"tprgdb/components/serialization"
)

type ICompressionStruct interface {
	Find(string) (uint64, error)
	Add(string) uint64
}

type IDecompressionStruct interface {
	Find(uint64) (string, error)
	Add(uint64, string)
}

type ISerializableCompression interface {
	ICompressionStruct
	serialization.ISerializable
}

type ISerializableDecompression interface {
	IDecompressionStruct
	serialization.ISerializable
}

type CompressionManager struct {
	compressionStruct   ISerializableCompression
	decompressionStruct ISerializableDecompression
	compressionEnabled  bool
}

func NewCompressionManager(compressionStruct ISerializableCompression, decompressionStruct ISerializableDecompression, compressionEnabled bool) *CompressionManager {
	return &CompressionManager{
		compressionStruct:   compressionStruct,
		decompressionStruct: decompressionStruct,
		compressionEnabled:  compressionEnabled,
	}
}

func (cm *CompressionManager) GetCompressed(key string) (uint64, error) {
	return cm.compressionStruct.Find(key)
}
func (cm *CompressionManager) GetDecompressed(value uint64) (string, error) {
	return cm.decompressionStruct.Find(value)
}
func (cm *CompressionManager) IsCompressionEnabled() bool {
	return cm.compressionEnabled
}

func (cm *CompressionManager) SerializeToFile(file *blockmanager.BlockedFile) error {
	err := cm.compressionStruct.SerializeToFile(file)
	if err != nil {
		return err
	}
	err = cm.decompressionStruct.SerializeToFile(file)
	if err != nil {
		return err
	}
	return nil
}
func (cm *CompressionManager) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	err := cm.compressionStruct.DeserializeFromFile(file)
	if err != nil {
		return err
	}
	err = cm.decompressionStruct.DeserializeFromFile(file)
	if err != nil {
		return err
	}
	return nil
}
func (cm *CompressionManager) SerializeToBytes() []byte {
	data := cm.compressionStruct.SerializeToBytes()
	data = append(data, cm.decompressionStruct.SerializeToBytes()...)
	return data
}
func (cm *CompressionManager) DeserializeFromBytes(data []byte) {
	cm.compressionStruct.DeserializeFromBytes(data)
	data = data[len(cm.compressionStruct.SerializeToBytes()):]
	cm.decompressionStruct.DeserializeFromBytes(data)
}

func (cm *CompressionManager) Add(key string) uint64 {
	compressedKey := cm.compressionStruct.Add(key)
	cm.decompressionStruct.Add(compressedKey, key)
	return compressedKey
}
