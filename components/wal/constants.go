package wal

const (
	// Fragment types
	FragmentTypePadding uint8 = 0
	FragmentTypeFull    uint8 = 1
	FragmentTypeFirst   uint8 = 2
	FragmentTypeMiddle  uint8 = 4
	FragmentTypeLast    uint8 = 8
	SegmentNumberSize   uint8 = 8
	SegmentSeedSize     uint8 = 4
	TypeSize            uint8 = 1
	TimestampSize       uint8 = 8
	TombstoneSize       uint8 = 1
	KeySize             uint8 = 8
	ValueSize           uint8 = 8
	CRCSize             uint8 = 8

	WALLeadingZeroes = 8
)

var HashSeed = []byte{0xAB, 0xCD, 0xEF, 0x12}
