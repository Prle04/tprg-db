package skiplist

import (
	"math/rand"
)

// type MemtableStruct interface {
// 	Set(key string, value []byte)
// 	Get(key string) []byte
// 	Del(key string)
// 	NumberOfKeys() uint64
// }

type SkipListNode struct {
	key       string
	data      []byte
	tombstone bool
	right     *SkipListNode
	down      *SkipListNode
	up        *SkipListNode
}

type SkipList struct {
	head  *SkipListNode
	level int
	size  uint64
}

func NewSkipList(levels int) *SkipList {
	if levels <= 0 {
		panic("Number of levels must be greater than 0")
	}

	head := &SkipListNode{}
	current := head
	for i := 1; i < levels; i++ {
		newHead := &SkipListNode{}
		current.down = newHead
		newHead.up = current
		current = newHead
	}

	return &SkipList{
		head:  head,
		level: levels,
		size:  0,
	}
}

func (sl *SkipList) randomLevel() int {

	level := 1
	for rand.Float32() < 0.5 && level < sl.level {
		level++
	}
	return level
}

func (sl *SkipList) Set(key string, value []byte) {
	update := make([]*SkipListNode, sl.level)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		update[i] = current
		if current.right != nil && current.right.key == key {
			current = current.right
			for current != nil {
				current.data = value
				current.tombstone = false
				current = current.down
			}
			return
		}
		if current.down != nil {
			current = current.down
		}
	}

	newLevel := sl.randomLevel()
	var downNode *SkipListNode
	for i := 0; i < newLevel; i++ {
		newNode := &SkipListNode{
			key:       key,
			data:      value,
			tombstone: false,
		}

		newNode.right = update[i].right
		update[i].right = newNode

		newNode.down = downNode
		if downNode != nil {
			downNode.up = newNode
		}

		downNode = newNode
	}

	sl.size++
}

func (sl *SkipList) Get(key string) []byte {
	current := sl.head

	for current != nil {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		if current.right != nil && current.right.key == key && !current.right.tombstone {
			return current.right.data
		}
		current = current.down
	}

	return nil
}

func (sl *SkipList) Del(key string) {
	current := sl.head

	for current != nil {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		if current.right != nil && current.right.key == key {
			current = current.right
			for current != nil {
				current.tombstone = true
				current = current.down
			}
			break

		}
		current = current.down
	}

	sl.size--
}

func (sl *SkipList) NumberOfKeys() uint64 {
	return sl.size
}

// SKIP LISTA ALI KADA SE NE ZNA UNAPRED KOLIKO NIVOA CE IMATI

// func NewSkipList() *SkipList {

// 	head := &SkipListNode{}

// 	return &SkipList{
// 		head:  head,
// 		level: 1,
// 		size:  0,
// 	}
// }

// func (sl *SkipList) randomLevel() int {

// 	level := 1
// 	for rand.Float32() < 0.5 {
// 		level++
// 	}
// 	return level
// }

// func (sl *SkipList) Set(key string, value []byte) {

// 	update := make([]*SkipListNode, sl.level)
// 	current := sl.head

// 	for i := sl.level - 1; i >= 0; i-- {
// 		for current.right != nil && current.right.key < key {
// 			current = current.right
// 		}
// 		update[i] = current
// 		if current.down != nil {
// 			current = current.down
// 		}
// 	}

// 	if current.right != nil && current.right.key == key {
// 		current = current.right
// 		for current != nil {
// 			current.data = value
// 			current.tombstone = false
// 			current = current.down
// 		}
// 		return
// 	}

// 	newLevel := sl.randomLevel()

// 	if newLevel > sl.level {
// 		for i := sl.level; i < newLevel; i++ {
// 			newHead := &SkipListNode{down: sl.head}
// 			sl.head.up = newHead
// 			sl.head = newHead
// 			update = append(update, sl.head)
// 		}
// 		sl.level = newLevel
// 	}

// 	var downNode *SkipListNode
// 	for i := 0; i < newLevel; i++ {
// 		newNode := &SkipListNode{
// 			key:       key,
// 			data:      value,
// 			tombstone: false,
// 			down:      downNode,
// 		}
// 		if downNode != nil {
// 			downNode.up = newNode
// 		}

// 		if i < len(update) && update[i] != nil {
// 			newNode.right = update[i].right
// 			update[i].right = newNode
// 		}

// 		downNode = newNode
// 	}

// 	sl.size++
// }

// func (sl *SkipList) Get(key string) []byte {
// 	current := sl.head

// 	for current != nil {
// 		for current.right != nil && current.right.key < key {
// 			current = current.right
// 		}
// 		if current.right != nil && current.right.key == key && !current.right.tombstone {
// 			return current.right.data
// 		}
// 		current = current.down
// 	}

// 	return nil
// }

// func (sl *SkipList) Del(key string) {
// 	current := sl.head

// 	for current != nil {
// 		for current.right != nil && current.right.key < key {
// 			current = current.right
// 		}
// 		if current.right != nil && current.right.key == key {
// 			current = current.right
// 			for current != nil {
// 				current.tombstone = true
// 				current = current.down
// 			}
// 			break

// 		}
// 		current = current.down
// 	}

// 	sl.size--
// }

// func (sl *SkipList) NumberOfKeys() uint64 {
// 	return sl.size
// }
