package memtable

import (
	"fmt"
	"tprgdb/components/btree"
)

type MemtableDataStructure interface {
	Set(key string, value []byte, isReadWrite bool)
	Delete(key string, isReadWrite bool)
	Get(key string) []byte
}

type Memtable struct {
	trees           []*btree.BTree // Niz B-stabala
	activeTreeIndex int            // Indeks aktivnog read-write stabla
	maxElements     uint32         // Maksimalan broj elemenata u read-write stablu pre prelaska na read-only
	wallSegmentData uint64
}

func NewMemtable(numTrees int, maxChildren, maxElements uint32) (*Memtable, error) {
	trees := make([]*btree.BTree, numTrees)
	for i := range trees {
		tree, err := btree.NewBTree(maxChildren)
		if err != nil {
			return nil, err
		}
		trees[i] = tree
	}

	return &Memtable{
		trees:           trees,
		activeTreeIndex: 0,
		maxElements:     maxElements,
	}, nil
}

func (m *Memtable) Set(key string, value []byte) {
	activeTree := m.trees[m.activeTreeIndex]
	activeTree.Set(key, value, true) // true označava da je u read-write modu

	if activeTree.GetNumberOfKeys() >= m.maxElements {
		// Prebacivanje trenutnog stabla u read-only mod
		fmt.Println("Switching active B-Tree to read-only mode")
		activeTree.SetReadWrite(false)

		// Prelazak na sledeće stablo
		m.activeTreeIndex = (m.activeTreeIndex + 1) % len(m.trees)
		m.trees[m.activeTreeIndex].SetReadWrite(true)
	}
}

func (m *Memtable) Delete(key string) {
	m.trees[m.activeTreeIndex].Delete(key, true) // true označava da je u read-write modu
}

func (m *Memtable) Get(key string) []byte {
	for _, tree := range m.trees {
		value := tree.Get(key)
		if value != nil {
			return value
		}
	}
	return nil
}
