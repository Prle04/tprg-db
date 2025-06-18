package merkletree

import (
	"math"
	"tprgdb/components/blockmanager"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type MerkleTree struct {
	nodes []uint64
}

func hashData(data *[]byte) uint64 {
	return hash.GetMD5Hash(*data)
}

func NewMerkleTree(data []*[]byte) *MerkleTree {
	if len(data) == 0 {
		return nil
	}
	hashes := make([]uint64, len(data))
	for i, d := range data {
		if d == nil {
			return nil
		} else {
			hashes[i] = hashData(d)
		}
	}
	return NewMerkleTreeFromHashes(hashes)
}

func NewMerkleTreeFromHashes(hashes []uint64) *MerkleTree {
	// number of elements
	n := len(hashes)
	if n == 0 {
		return nil
	}
	// height of the tree
	h := math.Ceil(math.Log2(float64(len(hashes))))
	// total number of nodes - number of all nodes in the tree expect the leaves plus the elements that will be leaves
	N := int(math.Pow(2, h)) - 1 + n
	// create the array of nodes
	nodes := make([]uint64, N)
	// fill the leaves of the tree
	for i := N - n; i < N; i++ {
		nodes[i] = hashes[i-N+n]
	}
	// fill the rest of the tree
	for i := N - n - 1; i >= 0; i-- {
		// left and right children
		l := i*2 + 1
		r := i*2 + 2
		// Data that will be hashed
		d := make([]byte, 0)
		if l < N {
			d = append(d, serialization.UInt64ToBytes(nodes[l])...)
		}
		if r < N {
			d = append(d, serialization.UInt64ToBytes(nodes[r])...)
		}
		// hash the data and store it in the current node
		nodes[i] = hash.GetMD5Hash(d)
	}
	return &MerkleTree{
		nodes: nodes,
	}
}

func (mt *MerkleTree) GetRoot() uint64 {
	if len(mt.nodes) == 0 {
		return 0
	}
	return mt.nodes[0]
}

func (mt *MerkleTree) SerializeToFile(file *blockmanager.BlockedFile) error {

	// Write the nodes
	for _, node := range mt.nodes {
		data := serialization.UInt64ToBytes(node)
		if err := file.Write(data); err != nil {
			return err
		}
	}
	// Write 0 to indicate the end of the nodes
	if err := file.Write(serialization.UInt64ToBytes(0)); err != nil {
		return err
	}
	return nil
}

func (mt *MerkleTree) DeserializeFromFile(file *blockmanager.BlockedFile) error {

	mt.nodes = make([]uint64, 0)
	// Read the nodes
	for {
		data, err := file.Read(8) // Each node is 8 bytes
		if err != nil {
			return err
		}
		value := serialization.BytesToUInt64(data)
		if value == 0 {
			break // End of nodes
		}
		mt.nodes = append(mt.nodes, value)
	}
	return nil
}
