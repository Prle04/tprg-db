package avltree_test

import (
	"fmt"
	"math/rand"
	"testing"
	"tprgdb/components/avltree"
)

func TestInsertAndFind(t *testing.T) {
	tree := avltree.CreateAVLTree()

	tree.Insert("key1", 10)
	tree.Insert("key2", 20)
	tree.Insert("key3", 30)

	node := tree.Find("key2")
	if node == nil || node.GetValue() != 20 {
		t.Errorf("Expected value 20 for key2, got %v", node)
	}

	// Непостојећи кључ
	missing := tree.Find("keyX")
	if missing != nil {
		t.Errorf("Expected nil for non-existent key, got %v", missing)
	}
}

func TestPrefixScan(t *testing.T) {
	tree := avltree.CreateAVLTree()
	keys := []string{"abc", "abd", "abe", "abz", "bcd", "bce"}
	for i, k := range keys {
		tree.Insert(k, uint64(i))
	}

	results := tree.PrefixScan("ab", 1, 3)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	if results[0].GetKey() != "abc" || results[2].GetKey() != "abe" {
		t.Errorf("Unexpected prefix scan results: %v", results)
	}

	// Друга страница
	results = tree.PrefixScan("ab", 2, 2)
	if len(results) != 2 || results[0].GetKey() != "abe" {
		t.Errorf("Unexpected paginated prefix scan results: %v", results)
	}
}

func TestCount(t *testing.T) {
	tree := avltree.CreateAVLTree()

	keys := make([]string, 100)
	for i := range 100 {
		keys[i] = fmt.Sprintf("key-%3d", i)
	}

	for i, key := range keys {
		tree.Insert(key, uint64(i))
	}

	if tree.GetNumberOfElements() != 100 {
		t.Fatal("Not enough keys")
	}

	rand.Shuffle(100, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	for i, key := range keys {
		tree.Insert(key, uint64(i+100))
	}

	if tree.GetNumberOfElements() != 100 {
		t.Fatal("Not enough keys")
	}

}

func TestRangeScan(t *testing.T) {
	tree := avltree.CreateAVLTree()

	// Ubacujemo podatke
	podaci := []struct {
		key   string
		value uint64
	}{
		{"apple", 1},
		{"banana", 2},
		{"cherry", 3},
		{"date", 4},
		{"fig", 5},
		{"grape", 6},
		{"kiwi", 7},
	}

	for _, p := range podaci {
		tree.Insert(p.key, p.value)
	}

	// Test 1: opseg od "banana" do "fig", strana 1, po 3 elementa
	result := tree.RangeScan("banana", "fig", 1, 3)

	expectedKeys := []string{"banana", "cherry", "date"}
	if len(result) != len(expectedKeys) {
		t.Errorf("Expected %d elements, got %d", len(expectedKeys), len(result))
	}

	for i, data := range result {
		if data.GetKey() != expectedKeys[i] {
			t.Errorf("Expected key %s at index %d, got %s", expectedKeys[i], i, data.GetKey())
		}
	}

	// Test 2: opseg od "banana" do "fig", strana 2, po 3 elementa
	result = tree.RangeScan("banana", "fig", 2, 3)

	expectedKeys = []string{"fig"}
	if len(result) != len(expectedKeys) {
		t.Errorf("Expected %d elements, got %d", len(expectedKeys), len(result))
	}

	for i, data := range result {
		if data.GetKey() != expectedKeys[i] {
			t.Errorf("Expected key %s at index %d, got %s", expectedKeys[i], i, data.GetKey())
		}
	}

	// Test 3: opseg od "apple" do "kiwi", strana 1, po 10 elemenata (svi)
	result = tree.RangeScan("apple", "kiwi", 1, 10)

	if len(result) != len(podaci) {
		t.Errorf("Expected %d elements, got %d", len(podaci), len(result))
	}

	for i, data := range result {
		if data.GetKey() != podaci[i].key {
			t.Errorf("Expected key %s at index %d, got %s", podaci[i].key, i, data.GetKey())
		}
	}

	fmt.Println("RangeScan testovi su prosli.")
}
