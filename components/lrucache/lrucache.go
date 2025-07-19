package lrucache

type LRUCache[T any] struct {
	maxSize     uint64
	currentSize uint64
	head        *LRUNode[T]
	tail        *LRUNode[T]
	cache       map[string]*LRUNode[T]
}

func NewLRUCache[T any](maxSize uint64) *LRUCache[T] {
	return &LRUCache[T]{
		maxSize:     maxSize,
		currentSize: 0,
		cache:       make(map[string]*LRUNode[T]),
	}
}

// Returns value and if the value was found in the cache
func (c *LRUCache[T]) GetElement(key string) (T, bool) {
	node, ok := c.cache[key]
	if !ok {
		// key is not in the cache, return zero value of T and false
		var zeroValue T
		return zeroValue, false
	}
	// Move the accessed node to the front of the list
	// If the node is already at the front, return its value
	if node == c.head {
		return node.GetValue(), true
	}
	// If the node is at the end of the list, update the tail
	if node == c.tail {
		c.tail = node.GetPrev()
		c.tail.SetNext(nil)
	}
	// Remove the node from its current position
	prev := node.GetPrev()
	next := node.GetNext()
	if prev != nil {
		prev.SetNext(next)
	}
	if next != nil {
		next.SetPrev(prev)
	}
	// Add the node to the front of the list
	node.SetPrev(nil)
	node.SetNext(c.head)
	if c.head != nil {
		c.head.SetPrev(node)
	}
	c.head = node
	return node.GetValue(), true
}

func (c *LRUCache[T]) AddElement(key string, value T) (bool, string) {

	// If the key is already in the cache, move it to the front of the list
	if elem, ok := c.cache[key]; ok {
		elem.SetValue(value) // Update the block pointer in the cache
		// If the block is already at the front of the list, return
		if elem == c.head {
			return false, ""
		}
		// Remove the block from the list
		elem.GetPrev().SetNext(elem.GetNext())
		if elem.GetNext() != nil {
			elem.GetNext().SetPrev(elem.GetPrev())
		}
		// If the block is at the end of the list, update the tail
		if elem == c.tail {
			elem.GetPrev().SetNext(nil)
			c.tail = elem.GetPrev()
		}
		// Add the block to the front of the list
		elem.SetPrev(nil)
		elem.SetNext(c.head)
		c.head.SetPrev(elem)
		c.head = elem
		return false, ""
	}
	// From this point on, we know the block is not in the cache and we need to add new block
	// and evict old block if necessary
	var evictionPath string = ""

	if c.currentSize == c.maxSize {
		evictionPath = c.evict()
	}
	c.currentSize++
	newNode := &LRUNode[T]{
		key:   key,
		value: value,
		next:  c.head,
		prev:  nil,
	}
	if c.head == nil {
		c.head = newNode
		c.tail = newNode
	} else {
		c.head.SetPrev(newNode)
		c.head = newNode
	}
	c.cache[key] = newNode
	return true, evictionPath
}

func (c *LRUCache[T]) evict() string {
	// If the cache is empty, return
	if c.head == nil {
		return ""
	}
	// There is only one block in the cache
	if c.head == c.tail {
		key := c.head.GetKey()
		delete(c.cache, key)
		c.head = nil
		c.tail = nil
		c.currentSize--
		return key
	}
	// Remove the block from the cache
	key := c.tail.GetKey()
	delete(c.cache, key)
	c.tail.GetPrev().SetNext(nil)
	c.tail = c.tail.GetPrev()
	c.currentSize--
	return key
}

func (c *LRUCache[T]) RemoveElement(key string) {
	if _, ok := c.cache[key]; !ok {
		return
	}
	node := c.cache[key]
	if node == c.head {
		c.head = node.GetNext()
	}
	if node == c.tail {
		c.tail = node.GetPrev()
	}
	if node.GetPrev() != nil {
		node.GetPrev().SetNext(node.GetNext())
	}
	if node.GetNext() != nil {
		node.GetNext().SetPrev(node.GetPrev())
	}
	delete(c.cache, key)
	c.currentSize--
}

func (c *LRUCache[T]) GetIterator() *LRUIterator[T] {
	return NewLRUIterator(c)
}

func (c *LRUCache[T]) Size() uint64 {
	return c.currentSize
}

func (c *LRUCache[T]) Exists(key string) bool {
	_, exists := c.cache[key]
	return exists
}