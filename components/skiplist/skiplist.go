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

/*
Strukture:
- SkipList: Glavna struktura koja sadrži :pokazivač na glavu (prazan Node koji je na
 najvecem nivou), trenutni nivo,
  veličinu (broj elemenata) i maksimalnu visinu (ako je negativna nema ogranicenja
  u suprotnom to ce biti granica/maksimalna visina).
- SkipListNode: Čvor u skip listi sa pokazivačima na susedne čvorove (desno,dole i gore),
  ključ, podatke i oznake za brisanje (tombstone).


  U zavisnosti da li postoji ogranicenje za maksimalnu visinu
  malo je razlicita logika tokom randomLevel funkcije i dodavanja novog elementa.

*/

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
	level      int
	size       uint64
	max_height int
}

func NewSkipList(max_height int) *SkipList {
	head := &SkipListNode{}
	level := 1

	return &SkipList{
		head:       head,
		level:      level,
		size:       0,
		max_height: max_height,
	}

}

func (sl *SkipList) randomLevel() int {

	level := 1
	if sl.max_height > 0 {
		for rand.Float32() < 0.5 && level <= sl.max_height {
			level++
		}
	} else {
		for rand.Float32() < 0.5 {
			level++
		}
	}

	return level
}

func (sl *SkipList) Adding(key string, value []byte, is_delited bool, update []*SkipListNode) {
	newLevel := sl.randomLevel()

	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			newHead := &SkipListNode{down: sl.head}
			sl.head.up = newHead
			sl.head = newHead
			update = append(update, sl.head)
		}
		sl.level = newLevel
	}

	var downNode *SkipListNode
	for i := 0; i < newLevel; i++ {
		newNode := &SkipListNode{
			key:       key,
			data:      value,
			tombstone: is_delited,
		}

		//povezivanje horizontalno
		if i < len(update) && update[i] != nil {
			newNode.right = update[i].right
			update[i].right = newNode
		}

		newNode.down = downNode
		if downNode != nil {
			downNode.up = newNode
		}

		downNode = newNode
	}

	sl.size++

}

func (sl *SkipList) Finding_and_updating(key string, value []byte, is_delited bool, is_update bool) ([]*SkipListNode, bool) {
	current := sl.head
	updateNodes := make([]*SkipListNode, sl.level)
	for i := sl.level - 1; i >= 0; i-- {
		for current.right != nil && current.right.key < key {
			current = current.right
		}
		updateNodes[i] = current
		if current.right != nil && current.right.key == key {
			current = current.right
			for current != nil {
				if is_update {
					current.data = value
				}
				current.tombstone = is_delited

				current = current.down
			}
			return updateNodes, true
		}
		if current.down != nil {
			current = current.down
		}
	}
	return updateNodes, false

}

func (sl *SkipList) Set(key string, value []byte) {
	update, found := sl.Finding_and_updating(key, value, false, true)

	if !found {

		sl.Adding(key, value, false, update)
	}

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
	update, found := sl.Finding_and_updating(key, nil, true, false)

	if !found {
		sl.Adding(key, nil, true, update)
	}

}

func (sl *SkipList) NumberOfKeys() uint64 {
	return sl.size
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
