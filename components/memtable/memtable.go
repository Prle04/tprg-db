package memtable

import "tprgdb/components/btree"
//niz instanci
//wall segment number

type MemtableDataStructure interface {
	Set(key string, value []byte)
	Delete(key string)
	Get(key string) (value []byte)
}

type Memtable struct {
	data MemtableDataStructure
	wallSegmentData uint64
}


func NewMemtable(maxChildren uint32) (*Memtable, error) {
	btree, err := btree.NewBTree(maxChildren)
	if err != nil {
		return nil, err
	}
	return &Memtable{
		data: btree,
	}, nil
}

func (m *Memtable) Set(key string, value []byte) {
	m.data.Set(key, value)
}

func (m *Memtable) Get(key string) []byte {
	return m.data.Get(key)
}








