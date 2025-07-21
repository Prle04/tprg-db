package memtable

import "tprgdb/components/skiplist"

type SkipListMemtableAdapter struct {
	skiplist *skiplist.SkipList
}

func NewSkipListMemtableAdapter() *SkipListMemtableAdapter {
	return &SkipListMemtableAdapter{
		skiplist: skiplist.NewSkipList(),
	}
}

func (s *SkipListMemtableAdapter) Get(key string) IMemtableItem {
	return s.skiplist.Get(key)
}
func (s *SkipListMemtableAdapter) Set(key string, value []byte, timestamp uint64) {

	s.skiplist.Set(key, value, timestamp)
}

func (s *SkipListMemtableAdapter) Delete(key string, timestamp uint64) {
	s.skiplist.Delete(key, timestamp)
}
func (s *SkipListMemtableAdapter) Clear() {
	s.skiplist.Clear()
}
func (s *SkipListMemtableAdapter) GetNumberOfKeys() int {
	return s.skiplist.GetNumberOfKeys()
}
func (s *SkipListMemtableAdapter) GetAllItems() []IMemtableItem {
	skipListItems := s.skiplist.GetAllItems()
	items := make([]IMemtableItem, len(skipListItems))
	for i, item := range skipListItems {
		items[i] = item
	}
	return items
}
