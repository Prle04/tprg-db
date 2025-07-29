package lrucache

type LRUIterator[T any] struct {
	current *LRUNode[T]
}

func NewLRUIterator[T any](cache *LRUCache[T]) *LRUIterator[T] {
	return &LRUIterator[T]{current: cache.head}
}

func (it *LRUIterator[T]) HasNext() bool {
	return it.current != nil
}

func (it *LRUIterator[T]) Next() (string, T, bool) {
	if it.current == nil {
		var zeroValue T
		return "", zeroValue, false
	}
	value := it.current.GetValue()
	key := it.current.GetKey()
	it.current = it.current.GetNext()
	return key, value, true
}
