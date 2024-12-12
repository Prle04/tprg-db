package lrucache


type LRUNode[T any] struct {
	key   string
	value T
	prev  *LRUNode[T]
	next  *LRUNode[T]
}


func (n *LRUNode[T]) GetNext() *LRUNode[T] {
	return n.next
}

func (n *LRUNode[T]) SetNext(next *LRUNode[T]) {
	n.next = next
}

func (n *LRUNode[T]) GetPrev() *LRUNode[T] {
	return n.prev
}

func (n *LRUNode[T]) SetPrev(prev *LRUNode[T]) {
	n.prev = prev
}

func (n *LRUNode[T]) GetKey() string {
	return n.key
}

func (n *LRUNode[T]) GetValue() T {
	return n.value
}

func (n *LRUNode[T]) SetValue(value T) {
	n.value = value
}