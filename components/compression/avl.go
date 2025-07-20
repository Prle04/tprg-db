package compression

import (
	"fmt"
	"tprgdb/components/avltree"
	"tprgdb/components/blockmanager"
)

// COMPRESSION STRUCTURE

type AVLCompression struct {
	tree *avltree.AVLTree
}

func NewAVLCompression(tree *avltree.AVLTree) *AVLCompression {
	return &AVLCompression{
		tree: tree,
	}
}

func (h *AVLCompression) Find(key string) (uint64, error) {
	value := h.tree.Find(key)
	if value != 0 {
		return value, nil
	}
	return 0, fmt.Errorf("key %s not found in compression dictionary", key)
}
func (h *AVLCompression) SerializeToBytes() []byte {
	return h.tree.SerializeToBytes()
}
func (h *AVLCompression) DeserializeFromBytes(data []byte) {
	h.tree.DeserializeFromBytes(data)
}
func (h *AVLCompression) SerializeToFile(file *blockmanager.BlockedFile) error {
	return h.tree.SerializeToFile(file)
}
func (h *AVLCompression) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	return h.tree.DeserializeFromFile(file)
}

func (h *AVLCompression) Add(key string) uint64 {
	value := h.tree.Find(key)
	if value != 0 {
		return value
	}
	compressedKey := uint64(h.tree.GetNumberOfElements() + 1) // Simple incremental compression
	h.tree.Insert(key, compressedKey)
	return compressedKey
}
