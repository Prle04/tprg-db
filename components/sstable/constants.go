package sstable

const (
	// Fragment types
	FragmentTypePadding uint8 = 0
	FragmentTypeFull    uint8 = 1
	FragmentTypeFirst   uint8 = 2
	FragmentTypeMiddle  uint8 = 4
	FragmentTypeLast    uint8 = 8

	SegmentNumberSize uint8 = 8
	TypeSize          uint8 = 1
	TombstoneSize     uint8 = 1
	CRCSize           uint8 = 8

	WALLeadingZeroes = 8

	SegmentBlockSize uint64 = 512
	PaddingThreshold uint64 = 32

	SSTableHeaderSize uint64 = 41 // Size of the SSTable header in bytes

	MODE_ONE_FILE   = "one-file"
	MODE_MULTI_FILE = "multi-file"
)
