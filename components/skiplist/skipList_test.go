package skiplist

import (
	"bytes"
	"testing"
	"time"
)

func TestSkipListSetAndGet(t *testing.T) {
	sl := NewSkipList()
	key := "foo"
	value := []byte("bar")

	sl.Set(key, value, uint64(time.Now().Unix()))
	node := sl.Get(key)

	if node == nil {
		t.Fatal("Expected node, got nil")
	}

	if node.GetKey() != key {
		t.Errorf("Expected key %s, got %s", key, node.GetKey())
	}

	if !bytes.Equal(node.GetValue(), value) {
		t.Errorf("Expected value %s, got %s", value, node.GetValue())
	}

	if node.IsDeleted() {
		t.Error("Expected node not to be marked as deleted")
	}
}

func TestSkipListDelete(t *testing.T) {
	sl := NewSkipList()
	key := "baz"
	value := []byte("qux")

	sl.Set(key, value, uint64(time.Now().Unix()))
	sl.Delete(key, uint64(time.Now().Unix()))

	node := sl.Get(key)
	if node == nil {
		t.Fatal("Expected node after delete, got nil")
	}
	if !node.IsDeleted() {
		t.Error("Expected node to be marked as deleted")
	}
}

func TestSkipListClear(t *testing.T) {
	sl := NewSkipList()
	sl.Set("a", []byte("1"), uint64(time.Now().Unix()))
	sl.Set("b", []byte("2"), uint64(time.Now().Unix()))

	sl.Clear()

	if sl.GetNumberOfKeys() != 0 {
		t.Errorf("Expected 0 keys, got %d", sl.GetNumberOfKeys())
	}

	if sl.head == nil {
		t.Error("Expected non-nil head after clear")
	}
}

func TestSkipListGetAllItems(t *testing.T) {
	sl := NewSkipList()
	sl.Set("a", []byte("1"), uint64(time.Now().Unix()))
	sl.Set("b", []byte("2"), uint64(time.Now().Unix()))
	sl.Set("c", []byte("3"), uint64(time.Now().Unix()))

	items := sl.GetAllItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	expectedKeys := []string{"a", "b", "c"}
	for i, item := range items {
		if item.GetKey() != expectedKeys[i] {
			t.Errorf("Expected key %s, got %s", expectedKeys[i], item.GetKey())
		}
	}
}

func TestSkipListUpdateValue(t *testing.T) {
	sl := NewSkipList()
	sl.Set("x", []byte("old"), uint64(time.Now().Unix()))
	sl.Set("x", []byte("new"), uint64(time.Now().Unix()))

	node := sl.Get("x")
	if node == nil {
		t.Fatal("Expected node, got nil")
	}
	if !bytes.Equal(node.GetValue(), []byte("new")) {
		t.Errorf("Expected updated value 'new', got %s", node.GetValue())
	}
}
