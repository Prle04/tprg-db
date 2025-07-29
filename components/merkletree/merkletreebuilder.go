package merkletree

type MerkleTreeBuilder struct {
	hashes []uint64
}

func NewMerkleTreeBuilder() *MerkleTreeBuilder {
	return &MerkleTreeBuilder{
		hashes: make([]uint64, 0),
	}
}

func (mtb *MerkleTreeBuilder) AddData(data *[]byte) {
	if data == nil {
		return
	}
	mtb.hashes = append(mtb.hashes, hashData(data))
}

func (mtb *MerkleTreeBuilder) Build() *MerkleTree {
	if len(mtb.hashes) == 0 {
		return nil
	}
	return NewMerkleTreeFromHashes(mtb.hashes)
}