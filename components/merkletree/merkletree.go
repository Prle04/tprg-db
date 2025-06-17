package merkletree

import (
	"math"
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
	// number of elements
	n := len(data)
	if data == nil || n == 0 {
		return nil
	}
	// height of the tree
	h := math.Ceil(math.Log2(float64(len(data))))
	// total number of nodes - number of all nodes in the tree expect the leaves plus the elements that will be leaves
	N := int(math.Pow(2, h)) - 1 + n
	// create the array of nodes
	nodes := make([]uint64, N)
	// fill the leaves of the tree
	for i := N - n; i < N; i++ {
		nodes[i] = hashData(data[i-N+n])
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
