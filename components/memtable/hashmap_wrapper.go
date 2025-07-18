package memtable

import (
	"sort"
	"time"
)

// type MemtableStruct interface {
// 	Set(key string, value []byte)
// 	Get(key string) []byte
// 	Del(key string)
// 	NumberOfKeys() uint64
// }

type element struct {
	key       string
	value     []byte
	timestamp uint64
	isDeleted bool
}

type HashMapWrapper struct {
	store map[string]*element
	size  int
}

func NewHashMapWrapper() *HashMapWrapper {
	return &HashMapWrapper{
		store: make(map[string]*element),
		size:  0,
	}
}

func (hm *HashMapWrapper) Set(key string, value []byte) {
	if el, exists := hm.store[key]; exists {
		el.isDeleted = false
		el.timestamp = uint64(time.Now().Unix())
		el.value = value
	} else {
		hm.store[key] = &element{
			key:       key,
			value:     value,
			isDeleted: false,
			timestamp: uint64(time.Now().Unix()),
		}
		hm.size++
	}
}

func (hm *HashMapWrapper) Get(key string) IMemtableItem {
	if el, exists := hm.store[key]; exists {
		return el
	}
	return nil
}

func (hm *HashMapWrapper) Delete(key string) {
	if el, exists := hm.store[key]; exists {
		el.isDeleted = true
		el.timestamp = uint64(time.Now().Unix())
	} else {
		hm.store[key] = &element{
			key: key,
			// value:     value,
			isDeleted: true,
			timestamp: uint64(time.Now().Unix()),
		}
		hm.size++
	}
}

func (hm *HashMapWrapper) GetNumberOfKeys() int {
	return hm.size
}

func (e *element) GetKey() string {
	return e.key
}
func (e *element) GetValue() []byte {
	return e.value
}
func (e *element) IsDeleted() bool {
	return e.isDeleted
}
func (e *element) GetTimestamp() uint64 {
	return e.timestamp
}

func (hm *HashMapWrapper) Clear() {
	hm.store = make(map[string]*element)
	hm.size = 0
}

func (hm *HashMapWrapper) GetAllItems() []IMemtableItem {
	var items []IMemtableItem
	for _, el := range hm.store {
		items = append(items, el)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetKey() < items[j].GetKey()
	})
	return items
}
