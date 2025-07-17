package sstable

import "tprgdb/components/serialization"

type SSTableHeader struct {
	IndexStartOffset       uint64
	SummaryStartOffset     uint64
	BloomFilterStartOffset uint64
	MerkleStartOffset      uint64
	NumberOfItems          uint64
	IsCompressionEnabled   bool
}

func (header *SSTableHeader) SerializeToBytes() []byte {
	// 8B - IndexStartOffset
	// 8B - SummaryStartOffset
	// 8B - BloomFilterStartOffset
	// 8B - MerkleStartOffset

	result := make([]byte, 0, SSTableHeaderSize)

	// IndexStartOffset
	result = append(result, serialization.UInt64ToBytes(header.IndexStartOffset)...)
	// SummaryStartOffset
	result = append(result, serialization.UInt64ToBytes(header.SummaryStartOffset)...)
	// BloomFilterStartOffset
	result = append(result, serialization.UInt64ToBytes(header.BloomFilterStartOffset)...)
	// MerkleStartOffset
	result = append(result, serialization.UInt64ToBytes(header.MerkleStartOffset)...)
	// NumberOfItems
	result = append(result, serialization.UInt64ToBytes(header.NumberOfItems)...)
	// IsCompressionEnabled
	result = append(result, serialization.BoolToBytes(header.IsCompressionEnabled)...)
	return result
}

func (header *SSTableHeader) DeserializeFromBytes(data []byte) {
	header.IndexStartOffset = serialization.BytesToUInt64(data[:8])
	header.SummaryStartOffset = serialization.BytesToUInt64(data[8:16])
	header.BloomFilterStartOffset = serialization.BytesToUInt64(data[16:24])
	header.MerkleStartOffset = serialization.BytesToUInt64(data[24:32])
	header.NumberOfItems = serialization.BytesToUInt64(data[32:40])
	header.IsCompressionEnabled = serialization.BytesToBool(data[40:41])
}
