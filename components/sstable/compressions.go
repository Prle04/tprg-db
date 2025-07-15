package sstable

type ICompressionManager interface {
	GetCompressed(string) (uint64, error)
	GetDecompressed(uint64) (string, error)
	IsCompressionEnabled() bool
}
