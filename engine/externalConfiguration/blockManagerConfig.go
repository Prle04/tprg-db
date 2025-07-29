package externalconfiguration

type BlockManagerConfig struct {
	BlockSize      string `json:"blockSize"`
	BlockSizeBytes int
	MaxElements    int `json:"maxElements"`
}

const (
	KB4  = "4KB"
	KB8  = "8KB"
	KB16 = "16KB"
)

func (b *BlockManagerConfig) Validate(dbName string) {
	validSizes := map[string]int{
		KB4:  4096,
		KB8:  8192,
		KB16: 16384,
	}
	if size, exists := validSizes[b.BlockSize]; exists {
		b.BlockSizeBytes = size
	} else {
		b.BlockSizeBytes = 4096 // Default to 4KB if invalid
		b.BlockSize = KB4
	}

	if b.MaxElements <= 0 {
		b.MaxElements = 1024 // Default max elements
	}
}
