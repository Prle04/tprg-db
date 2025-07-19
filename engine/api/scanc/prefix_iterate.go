package scanc

import (
	"tprgdb/engine"
	"tprgdb/engine/core"
)

type PrefixIterator struct {
	rangeIterator *RangeIterator
}

func NewPrefixIterator(core *core.Core, prefix string) *PrefixIterator {
	return &PrefixIterator{
		rangeIterator: NewRangeIterator(core, prefix, prefix+"\uffff"),
	}
}

func (it *PrefixIterator) Next() (*engine.FinalItem, error) {
	return it.rangeIterator.Next()
}

func (it *PrefixIterator) Stop() {
	it.rangeIterator.Stop()
}
