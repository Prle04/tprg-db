package hashmapwrapper

import (
	"sort"
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
	isDeleted bool
}

type HashMapWrapper struct {
	store map[string]*element
	size  uint64
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
		el.value = value
	} else {
		hm.store[key] = &element{
			key:       key,
			value:     value,
			isDeleted: false,
		}
		hm.size++
	}
}

func (hm *HashMapWrapper) Get(key string) *element {
	if el, exists := hm.store[key]; exists {
		return el
	}
	return nil
}

func (hm *HashMapWrapper) Delete(key string) {
	if el, exists := hm.store[key]; exists {
		el.isDeleted = true
	} else {
		hm.store[key] = &element{
			key: key,
			// value:     value,
			isDeleted: true,
		}
		hm.size++
	}
}

func (hm *HashMapWrapper) GetNumberOfKeys() uint64 {
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

func (hm *HashMapWrapper) Clear() {
	hm.store = make(map[string]*element)
	hm.size = 0
}

func (hm *HashMapWrapper) GetAllItems() []*element {
	var items []*element
	for _, el := range hm.store {
		items = append(items, el)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].key < items[j].key
	})
	return items
}
