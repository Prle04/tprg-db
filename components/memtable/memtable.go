package memtable

import (
	"errors"
)

// MemtableItem represents an item in the memtable with a key, value, and deletion status.
type IMemtableItem interface {
	GetKey() string
	GetValue() []byte
	GetTimestamp() uint64
	IsDeleted() bool
}

// MemtableDataStructure defines the interface for a data structure used in the memtable.
// So far implementated: SkipList, HashMapWrapper, and a Btree.
type IMemtableDataStructure interface {
	Get(key string) IMemtableItem
	Set(key string, value []byte, timestamp uint64)
	Delete(key string, timestamp uint64)

	Clear()

	GetNumberOfKeys() int

	// GetAllItems returns all items in the data structure sorted by key.
	GetAllItems() []IMemtableItem
}

type memtableStruct struct {
	dataStructureInstance IMemtableDataStructure
	// Contains a boolean indicating whether the instance needs to be flushed.
	flush bool
	// Contains the number of the last WAL segment that was
	lastSegmentNumber uint64
}

type Memtable struct {
	instances []memtableStruct

	// number of keys each instance can hold before it needs to be flushed.
	threshold         int
	numberOfInstances int

	// currentInstance is the index of the current instance being used for operations.
	currentInstance int
}

// type of a function that creates a new instance of MemtableDataStructure.
type NewInstanceFunc func() (IMemtableDataStructure, error)

func NewMemtable(
	numberOfInstances int,
	threshold int,
	newInstance NewInstanceFunc,
) (*Memtable, error) {
	if numberOfInstances <= 0 {
		return nil, errors.New("number of instances must be greater than zero")
	}

	// Create instances of MemtableDataStructure
	instances := make([]memtableStruct, numberOfInstances)
	for i := range numberOfInstances {
		dataStruct, err := newInstance()
		if err != nil {
			return nil, err
		}
		instances[i] = memtableStruct{
			dataStructureInstance: dataStruct,
			flush:                 false,
			lastSegmentNumber:     0,
		}
	}

	return &Memtable{
		instances:         instances,
		threshold:         threshold,
		numberOfInstances: numberOfInstances,

		currentInstance: 0,
	}, nil
}

func (m *Memtable) Get(key string) IMemtableItem {

	// Check current instance first
	item := m.instances[m.currentInstance].dataStructureInstance.Get(key)
	if item != nil {
		return item
	}

	// Loop through all instances to find the key
	// Starting from the previous instance
	// and going full circle
	index := (m.currentInstance - 1 + m.numberOfInstances) % m.numberOfInstances
	for index != m.currentInstance {
		item = m.instances[index].dataStructureInstance.Get(key)
		if item != nil {
			return item
		}
		// Move to the previous instance
		index = (index - 1 + m.numberOfInstances) % m.numberOfInstances
	}
	return nil // Key not found in any instance
}

func (m *Memtable) Set(key string, value []byte, timestamp uint64, segmentNumber uint64) error {
	// Check if the current instance has reached the threshold
	if m.instances[m.currentInstance].dataStructureInstance.GetNumberOfKeys() >= m.threshold {
		m.moveToNextInstance()
	}
	m.instances[m.currentInstance].dataStructureInstance.Set(key, value, timestamp)
	m.instances[m.currentInstance].lastSegmentNumber = segmentNumber
	return nil
}

func (m *Memtable) Delete(key string, timestamp uint64, segmentNumber uint64) error {
	// Check if the current instance has reached the threshold
	if m.instances[m.currentInstance].dataStructureInstance.GetNumberOfKeys() >= m.threshold {
		m.moveToNextInstance()
	}
	m.instances[m.currentInstance].dataStructureInstance.Delete(key, timestamp)
	m.instances[m.currentInstance].lastSegmentNumber = segmentNumber
	return nil
}

func (m *Memtable) moveToNextInstance() {
	m.instances[m.currentInstance].flush = true // Mark the current instance for flushing
	m.currentInstance = (m.currentInstance + 1) % m.numberOfInstances
	m.instances[m.currentInstance].dataStructureInstance.Clear()
}

func (m *Memtable) Flush() ([]IMemtableItem, uint64) {
	// Loop through all instances to see if they need to be flushed
	// Starting from the previous instance since the current instance cannot be flushed
	// and going full circle
	index := (m.currentInstance - 1 + m.numberOfInstances) % m.numberOfInstances
	for index != m.currentInstance {
		if m.instances[index].flush {
			m.instances[index].flush = false // Reset the flush flag
			return m.instances[index].dataStructureInstance.GetAllItems(),
				m.instances[index].lastSegmentNumber
		}
		// Move to the previous instance
		index = (index - 1 + m.numberOfInstances) % m.numberOfInstances
	}
	return nil, 0 // No instance needs to be flushed
}
