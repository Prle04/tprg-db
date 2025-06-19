package serialization

import (
	"tprgdb/components/blockmanager"
)

type Serializable interface {
	SerializeToBytes() []byte
	DeserializeFromBytes(data []byte)
	SerializeToFile(file *blockmanager.BlockedFile) error
	DeserializeFromFile(file *blockmanager.BlockedFile) error
}
