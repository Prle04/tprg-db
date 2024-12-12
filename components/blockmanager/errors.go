package blockmanager

type BlockOverflow struct{}
type BlockOutOfRange struct{}
type InvalidBlockPath struct{}

func (e *BlockOverflow) Error() string {
	return "there is not enough space in the block to write the data"
}

func (e *BlockOutOfRange) Error() string {
	return "the requested data is out of range"
}

func (e *InvalidBlockPath) Error() string {
	return "the block path is invalid"
}
