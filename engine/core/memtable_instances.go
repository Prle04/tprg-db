package core

import (
	"fmt"
	"tprgdb/components/memtable"
	externalconfiguration "tprgdb/engine/externalConfiguration"
)

func CreateInstance(config *externalconfiguration.MemtableConfig) memtable.NewInstanceFunc {
	switch config.DataStructure {
	case externalconfiguration.SKIPLIST:
		return func() (memtable.IMemtableDataStructure, error) {
			return memtable.NewSkipListMemtableAdapter(), nil
		}
	case externalconfiguration.HASHMAP:
		return func() (memtable.IMemtableDataStructure, error) {
			return memtable.NewHashMapWrapper(), nil
		}
	case externalconfiguration.BTREE:
		return func() (memtable.IMemtableDataStructure, error) {
			return memtable.NewBTreeWrapper(config)
		}

	}
	return func() (memtable.IMemtableDataStructure, error) {
		return nil, fmt.Errorf("unknown memtable structure: %s", config.DataStructure)
	}
}
