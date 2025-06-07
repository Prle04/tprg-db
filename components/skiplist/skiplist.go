package skiplist

import (
	"math/rand"
)

type SkipListNode struct {
	key       string
	data      []byte
	tombstone bool
	right     *SkipListNode
	down      *SkipListNode
	up        *SkipListNode
}

type SkipList struct {
	head       *SkipListNode
	size       int
	max_height int
}

func NewSkipList() *SkipList {
	head := &SkipListNode{}
	level := 1
	return &SkipList{
		head:       head,
		size:       0,
		max_height: level,
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for rand.Float64() < 0.5 && level < sl.max_height {
		level++
	}
	if level > sl.max_height {
		sl.max_height = level
	}
	return level
}

// Get retrieves a node by key. If the key exists, it returns the node; otherwise, it returns nil.
// It traverses the skip list from the head node down to the bottom level, checking each node's right pointer.

func (sl *SkipList) Get(key string) *SkipListNode {
	current := sl.head
	for current != nil {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		if current.right != nil && current.right.key == key {
			return current.right
		}
		current = current.down
	}
	return nil
}

// updatingSkipList is a helper function that updates the skip list with a new key-value pair or marks a key as deleted.
// It handles both insertion and deletion by checking if the key already exists.
// If the key exists, it updates the value and marks it as a tombstone if necessary.
// If the key does not exist, it creates a new node on random levels(its height) and inserts it into the skip list with the appropriate links and tombstone status.
// It also manages the height of the skip list, ensuring that it grows as needed.

func (sl *SkipList) updatingSkipList(key string, value []byte, tombstone bool) {
	current := sl.head
	// remember the nodes to update at each level
	updateNodes := make([]*SkipListNode, sl.max_height)
	for i := sl.max_height - 1; i >= 0; i-- {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		updateNodes[i] = current
		// if we find the key, we update it and return
		if current.right != nil && current.right.key == key {
			current = current.right
			if value == nil {
				value = current.data
			}
			for current != nil {
				current.data = value
				current.tombstone = tombstone
				current = current.down
			}
			return
		}
		// move down to the next level
		current = current.down
	}
	// if we reach here, the key does not exist, so we need to insert it, we need to create at which last level we will insert the new node
	newLevel := sl.randomLevel()
	if newLevel > sl.max_height {
		// if the new level is greater than the current max height, we need to update the head of the skip list
		// and create new head nodes for the new levels, and inseerting them in the updateNodes slice
		for i := sl.max_height; i < newLevel; i++ {
			newHead := &SkipListNode{down: sl.head}
			sl.head.up = newHead
			sl.head = newHead
			updateNodes = append(updateNodes, sl.head)
		}
		sl.max_height = newLevel
	}
	var downNode *SkipListNode
	// create the new node at each level
	for i := 0; i < newLevel; i++ {
		newNode := &SkipListNode{
			key:       key,
			data:      value,
			tombstone: tombstone,
		}
		// linking the horizontal pointers
		if i < len(updateNodes) && updateNodes[i] != nil {
			newNode.right = updateNodes[i].right
			updateNodes[i].right = newNode
		}
		// linking the vertical pointers of the same element
		newNode.down = downNode
		if downNode != nil {
			downNode.up = newNode
		}
		downNode = newNode
	}
	// if we are inserting a new node, we need to update the size of the skip list
	sl.size++
}

// Set adds a new key-value pair to the skip list or updates an existing key with a new value.
func (sl *SkipList) Set(key string, value []byte) {
	sl.updatingSkipList(key, value, false)
}

// Delete marks a key as deleted in the skip list by setting its tombstone flag to true if the key doesnt exist it adds it and marks it as deleted.
func (sl *SkipList) Delete(key string) {
	sl.updatingSkipList(key, nil, true)
}

// Clear resets the skip list, removing all nodes and resetting the size and max height.
func (sl *SkipList) Clear() {
	sl.head = &SkipListNode{}
	sl.size = 0
	sl.max_height = 1
}
func (sl *SkipList) GetNumberOfKeys() int {
	return sl.size
}

// GetAllItems returns all items in the skip list as a slice of SkipListNode pointers.
func (sl *SkipList) GetAllItems() []*SkipListNode {
	items := make([]*SkipListNode, 0, sl.size)
	current := sl.head
	for current.down != nil {
		current = current.down
	}
	// traverse the bottom level of the skip list and moving form the head to the first node containing data
	current = current.right
	for current != nil {
		items = append(items, current)
		current = current.right
	}
	return items
}

func (sl *SkipList) Print() {
	current := sl.head
	for current != nil {
		temp := current
		for temp != nil {
			if temp.tombstone {
				print("X")
			} else {
				print(temp.key)
			}
			print(" ")
			temp = temp.right
		}
		print("\n")
		current = current.down
	}
}

func (sn *SkipListNode) GetKey() string {
	return sn.key
}

func (sn *SkipListNode) GetValue() []byte {
	return sn.data
}
func (sn *SkipListNode) IsDeleted() bool {
	return sn.tombstone
}
