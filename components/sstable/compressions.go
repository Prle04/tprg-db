package sstable

type ICompressionManager interface {
	GetCompressied(string) ([]byte, error)
	GetDecompressied(uint64) (string, error)
	IsCompressionEnabled() bool
}
