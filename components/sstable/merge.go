package sstable

import (
	"container/heap"
	"tprgdb/components/blockmanager"
)

type mergeItem struct {
	Item        ISSTableItem
	ReaderIndex int
}

type mergeItemHeap struct {
	items                []mergeItem
	isCompressionEnabled bool
}

func (h mergeItemHeap) Len() int { return len(h.items) }
func (h mergeItemHeap) Less(i, j int) bool {

	if h.items[i].Item.GetKey() < h.items[j].Item.GetKey() {
		return true
	}
	if h.items[i].Item.GetKey() > h.items[j].Item.GetKey() {
		return false
	}
	return h.items[i].Item.GetTimestamp() > h.items[j].Item.GetTimestamp() // If keys are equal, compare by timestamp
}
func (h mergeItemHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *mergeItemHeap) Push(x interface{}) {
	h.items = append(h.items, x.(mergeItem))
}

func (h *mergeItemHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	return item
}

type SSTableMergeIterator struct {
	// readers       []*SSTableReader
	iterators     []ISSTableIterator
	lastItem      ISSTableItem
	priorityQueue *mergeItemHeap
}

func NewSSTableMergeIterator(readers []*SSTableReader, isCompressionEnabled bool) (ISSTableIterator, error) {
	if len(readers) == 0 {
		return nil, nil // No readers to merge
	}
	if len(readers) == 1 {
		return readers[0].GetIterator(), nil // If only one reader, return its iterator
	}

	iterators := make([]ISSTableIterator, len(readers))
	priorityQueue := &mergeItemHeap{
		items:                make([]mergeItem, 0, len(readers)),
		isCompressionEnabled: isCompressionEnabled,
	}
	for i, reader := range readers {
		var err error
		iterators[i] = reader.GetIterator()
		// Initialize the priority queue with the first item from each iterator
		item, err := iterators[i].Next()
		if err != nil {
			if err != blockmanager.EOF {
				return nil, err // Error reading first item
			}
			continue // If EOF, skip this reader
		}
		priorityQueue.items = append(priorityQueue.items, mergeItem{
			Item:        item,
			ReaderIndex: i,
		})
	}
	heap.Init(priorityQueue)

	return &SSTableMergeIterator{
		// readers:       readers,
		lastItem:      nil,
		iterators:     iterators,
		priorityQueue: priorityQueue,
	}, nil
}

func (m *SSTableMergeIterator) Next() (ISSTableItem, error) {
	if len(m.priorityQueue.items) == 0 {
		return nil, blockmanager.EOF // No more items to read
	}
	item := heap.Pop(m.priorityQueue).(mergeItem)
	// Read the next item from the same reader
	nextItem, err := m.iterators[item.ReaderIndex].Next()
	if err != nil {
		if err != blockmanager.EOF {
			return nil, err // Error reading next item
		}
	} else {
		// If we successfully read the next item, add it to the priority queue
		heap.Push(m.priorityQueue, mergeItem{
			Item:        nextItem,
			ReaderIndex: item.ReaderIndex,
		})
	}
	if m.lastItem != nil && item.Item.GetKey() == m.lastItem.GetKey() {
		return m.Next() // Skip duplicates
	}
	m.lastItem = item.Item
	return item.Item, nil
}
