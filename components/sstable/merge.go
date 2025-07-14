package sstable

import (
	"container/heap"
	"tprgdb/components/blockmanager"
)

type mergeItem struct {
	Item        ISSTableItem
	ReaderIndex int
}

type mergeItemHeap []mergeItem

func (h mergeItemHeap) Len() int { return len(h) }
func (h mergeItemHeap) Less(i, j int) bool {
	if h[i].Item.GetKey() < h[j].Item.GetKey() {
		return true
	}
	if h[i].Item.GetKey() > h[j].Item.GetKey() {
		return false
	}
	return h[i].Item.GetTimestamp() > h[j].Item.GetTimestamp() // If keys are equal, compare by timestamp
}
func (h mergeItemHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *mergeItemHeap) Push(x interface{}) {
	*h = append(*h, x.(mergeItem))
}

func (h *mergeItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

type SSTableMergeIterator struct {
	readers       []*SSTableReader
	iterators     []ISSTableIterator
	lastItem      ISSTableItem
	priorityQueue mergeItemHeap
}

func NewSSTableMergeIterator(readers []*SSTableReader) (ISSTableIterator, error) {
	if len(readers) == 0 {
		return nil, nil // No readers to merge
	}
	if len(readers) == 1 {
		return readers[0].GetIterator() // If only one reader, return its iterator
	}
	iterators := make([]ISSTableIterator, len(readers))
	priorityQueue := make(mergeItemHeap, 0, len(readers))
	for i, reader := range readers {
		var err error
		iterators[i], err = reader.GetIterator()
		if err != nil {
			return nil, err // Error getting iterator for reader
		}
		// Initialize the priority queue with the first item from each iterator
		item, err := iterators[i].Next()
		if err != nil {
			if err != blockmanager.EOF {
				return nil, err // Error reading first item
			}
			continue // If EOF, skip this reader
		}
		priorityQueue = append(priorityQueue, mergeItem{
			Item:        item,
			ReaderIndex: i,
		})
	}

	heap.Init(&priorityQueue)

	return &SSTableMergeIterator{
		readers:       readers,
		lastItem:      nil,
		iterators:     iterators,
		priorityQueue: priorityQueue,
	}, nil
}

func (m SSTableMergeIterator) Next() (ISSTableItem, error) {
	if len(m.priorityQueue) == 0 {
		return nil, blockmanager.EOF // No more items to read
	}
	item := heap.Pop(&m.priorityQueue).(mergeItem)
	if m.lastItem != nil && item.Item.GetKey() == m.lastItem.GetKey() {
		// If the item is a duplicate, we skip it
		return m.Next() // Recursively call Next to get the next unique item
	}
	m.lastItem = item.Item
	// Read the next item from the same reader
	nextItem, err := m.iterators[item.ReaderIndex].Next()
	if err != nil {
		if err != blockmanager.EOF {
			return nil, err // Error reading next item
		}
	} else {
		// If we successfully read the next item, add it to the priority queue
		heap.Push(&m.priorityQueue, mergeItem{
			Item:        nextItem,
			ReaderIndex: item.ReaderIndex,
		})
	}

	return nextItem, nil
}
