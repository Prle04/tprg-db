package scans

import (
	"tprgdb/engine/core"
	"tprgdb/engine/item"
)

type PrefixIterator struct {
	rangeIterator *RangeIterator
}

func NewPrefixIterator(core *core.Core, prefix string) *PrefixIterator {
	return &PrefixIterator{
		rangeIterator: NewRangeIterator(core, prefix, prefix+"\uffff"),
	}
}

func (it *PrefixIterator) Next() (*item.DBItem, error) {
	return it.rangeIterator.Next()
}

func (it *PrefixIterator) Stop() {
	it.rangeIterator.Stop()
}
