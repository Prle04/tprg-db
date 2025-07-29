package memtable

import (
	"tprgdb/components/btree"
	externalconfiguration "tprgdb/engine/externalConfiguration"
)

type BTreeWrapper struct {
	btree *btree.BTree
}

func NewBTreeWrapper(config *externalconfiguration.MemtableConfig) (*BTreeWrapper, error) {
	tree, err := btree.NewBTree(uint32(config.BTree.MaxChildren))
	if err != nil {
		return nil, err
	}
	return &BTreeWrapper{
		btree: tree,
	}, nil
}

func (b *BTreeWrapper) Get(key string) IMemtableItem {
	return b.btree.Get(key)
}
func (b *BTreeWrapper) Set(key string, value []byte, timestamp uint64) {
	b.btree.Set(key, value, timestamp)
}
func (b *BTreeWrapper) Delete(key string, timestamp uint64) {
	b.btree.Delete(key, timestamp)
}
func (b *BTreeWrapper) Clear() {
	b.btree.Clear()
}
func (b *BTreeWrapper) GetNumberOfKeys() int {
	return int(b.btree.GetNumberOfKeys())
}
func (b *BTreeWrapper) GetAllItems() []IMemtableItem {
	skipListItems := b.btree.GetAllItems()
	items := make([]IMemtableItem, len(skipListItems))
	for i, item := range skipListItems {
		items[i] = item
	}
	return items
}
