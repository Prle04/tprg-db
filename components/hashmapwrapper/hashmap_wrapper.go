package hashmapwrapper

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

func (hm *HashMapWrapper) NumberOfKeys() uint64 {
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
