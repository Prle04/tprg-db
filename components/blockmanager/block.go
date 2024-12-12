package blockmanager

// Block represents a block of data in the block manager.
type Block struct {
	size        uint64
	data        []byte
	isModified  bool
	orderNumber uint64
	filename    string
}

// NewBlock creates a new block with the given data.
func NewBlock(size uint64, filename string) *Block {
	return &Block{
		size:       size,
		data:       make([]byte, size),
		isModified: true,
		filename:   filename,
	}
}

func (b *Block) GetOrderNumber() uint64 {
	return b.orderNumber
}

func (b *Block) Size() uint64 {
	return b.size
}

func (b *Block) GetPath() string {
	return getBlockPath(b.filename, b.orderNumber)
}

func (b *Block) SetOrderNumber(orderNumber uint64) {
	b.orderNumber = orderNumber
}

func (b *Block) IsModified() bool {
	return b.isModified
}

func (b *Block) getRemainingSize(start uint64) uint64 {
	return b.size - start
}

func (b *Block) Write(start uint64, data []byte) error {
	if start+uint64(len(data)) > b.size {
		return &BlockOverflow{}
	}
	copy(b.data[start:], data)
	b.isModified = true
	return nil
}

func (b *Block) Read(start uint64, length uint64) ([]byte, error) {
	if start+length > b.size {
		return nil, &BlockOverflow{}
	}
	return b.data[start : start+length], nil
}
