package lrucache_test

import (
	"testing"
	"tprgdb/components/lrucache"
)

func TestCache(t *testing.T) {
	cache := lrucache.NewLRUCache[int](3)

	// Test adding elements
	if added, _ := cache.AddElement("a", 1); !added {
		t.Error("Expected to add element 'a'")
	}
	if added, _ := cache.AddElement("b", 2); !added {
		t.Error("Expected to add element 'b'")
	}
	if added, _ := cache.AddElement("c", 3); !added {
		t.Error("Expected to add element 'c'")
	}

	// Test that the cache size is correct
	if size := cache.Size(); size != 3 {
		t.Errorf("Expected cache size to be 3, got %d", size)
	}
	// Iterate through the cache
	iterator := cache.GetIterator()
	key, value, ok := iterator.Next()
	if !ok || key != "c" || value != 3 {
		t.Errorf("Expected to find element 'c' with value 3, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "b" || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "a" || value != 1 {
		t.Errorf("Expected to find element 'a' with value 1, got %s: %d", key, value)
	}
	// Test that there are no more elements
	if ok = iterator.HasNext(); ok {
		t.Error("Expected no more elements in the iterator")
	}

	// Test getting elements
	if value, found := cache.GetElement("a"); !found || value != 1 {
		t.Errorf("Expected to find element 'a' with value 1, got %v", value)
	}
	if value, found := cache.GetElement("b"); !found || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %v", value)
	}

	// Test itaring through the cache again
	iterator = cache.GetIterator()
	key, value, ok = iterator.Next()
	if !ok || key != "b" || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "a" || value != 1 {
		t.Errorf("Expected to find element 'a' with value 1, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "c" || value != 3 {
		t.Errorf("Expected to find element 'c' with value 3, got %s: %d", key, value)
	}
	// Test that there are no more elements
	if ok = iterator.HasNext(); ok {
		t.Error("Expected no more elements in the iterator")
	}

	// Test adding a new element that exceeds the max size
	if added, evicted := cache.AddElement("d", 4); !added {
		t.Error("Expected to add element 'd'")
	} else if evicted != "c" {
		t.Errorf("Expected to evict element 'c', got %s", evicted)
	}

	// Test that the cache size is still correct
	if size := cache.Size(); size != 3 {
		t.Errorf("Expected cache size to be 3 after eviction, got %d", size)
	}

	// Iterate through the cache again
	iterator = cache.GetIterator()
	key, value, ok = iterator.Next()
	if !ok || key != "d" || value != 4 {
		t.Errorf("Expected to find element 'd' with value 4, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "b" || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "a" || value != 1 {
		t.Errorf("Expected to find element 'a' with value 1, got %s: %d", key, value)
	}
	// Test that there are no more elements
	if ok = iterator.HasNext(); ok {
		t.Error("Expected no more elements in the iterator")
	}

	// Test that the least recently used element ('c') was evicted
	if _, found := cache.GetElement("c"); found {
		t.Error("Expected element 'c' to be evicted")
	}

	// Test that the cache still contains 'a' and 'b'
	if value, found := cache.GetElement("a"); !found || value != 1 {
		t.Errorf("Expected to find element 'a' with value 1, got %v", value)
	}
	if value, found := cache.GetElement("b"); !found || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %v", value)
	}

	// Test removing an element
	cache.RemoveElement("a")
	if _, found := cache.GetElement("a"); found {
		t.Error("Expected element 'a' to be removed")
	}
	// Test that 'b' is still present
	if value, found := cache.GetElement("b"); !found || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %v", value)
	}

	// Test removing an element that does not exist
	cache.RemoveElement("nonexistent")
	// Test that the cache size is still correct
	if size := cache.Size(); size != 2 {
		t.Errorf("Expected cache size to be 2 after removing 'a', got %d", size)
	}

	// Iterate through the cache again
	iterator = cache.GetIterator()
	key, value, ok = iterator.Next()
	if !ok || key != "b" || value != 2 {
		t.Errorf("Expected to find element 'b' with value 2, got %s: %d", key, value)
	}
	key, value, ok = iterator.Next()
	if !ok || key != "d" || value != 4 {
		t.Errorf("Expected to find element 'd' with value 4, got %s: %d", key, value)
	}
	_, _, ok = iterator.Next()
	if ok {
		t.Error("Expected no more elements in the iterator after removing 'a'")
	}
	if ok = iterator.HasNext(); ok {
		t.Error("Expected no more elements in the iterator after removing 'a'")
	}
}
