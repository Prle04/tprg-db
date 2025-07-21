package scanc

import (
	"errors"
	"tprgdb/engine/core"
	"tprgdb/engine/item"
)

type RangeIterator struct {
	rangeScanCursor *RangeScanCursor
	buffer          []*item.DBItem
	currentIndex    int
	stopped         bool
}

func NewRangeIterator(core *core.Core, lowerBound, upperBound string) *RangeIterator {
	return &RangeIterator{
		rangeScanCursor: NewRangeScanCursor(core, lowerBound, upperBound, ITERATOR_BUFFER_SIZE),
		buffer:          make([]*item.DBItem, 0, ITERATOR_BUFFER_SIZE),
		currentIndex:    -1,
	}
}

func (it *RangeIterator) Next() (*item.DBItem, error) {
	if it.stopped {
		return nil, errors.New("iterator has been stopped") // Iterator has been stopped
	}

	if it.currentIndex >= len(it.buffer) || it.rangeScanCursor.GetCurrentPage() == 0 {
		items, err := it.rangeScanCursor.GetPage(it.rangeScanCursor.GetCurrentPage() + 1)
		if err != nil {
			return nil, err
		}
		it.buffer = items
		it.currentIndex = 0

		if len(it.buffer) == 0 {
			it.stopped = true
			return nil, END
		}
	}

	item := it.buffer[it.currentIndex]
	it.currentIndex++
	return item, nil
}

func (it *RangeIterator) Stop() {
	it.stopped = true
	it.buffer = nil // Clear the buffer to free memory
}
