package serialization

import "os"

type Serializable interface {
	SerializeToBytes() []byte
	DeserializeFromBytes(data []byte)
	SerializeToFile(file os.File) error
	DeserializeFromFile(file os.File) error
}
