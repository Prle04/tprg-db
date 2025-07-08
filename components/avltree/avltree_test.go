package avltree_test

import (
	"fmt"
	"testing"
	"tprgdb/components/avltree"
)

func TestAVLInsertAndTraversal(t *testing.T) {
	tree := avltree.CreateAVLTree()

	// Убацујемо елементе
	tree.Insert("delta", 4)
	tree.Insert("alpha", 1)
	tree.Insert("epsilon", 5)
	tree.Insert("beta", 2)
	tree.Insert("gamma", 3)

	// Очекује се да је стабло балансирано, нека је само за визуелну проверу
	fmt.Print("InOrder:   ")
	tree.InOrder()

	fmt.Print("PreOrder:  ")
	tree.PreOrder()

	fmt.Print("PostOrder: ")
	tree.PostOrder()
}

func TestPrefixScan(t *testing.T) {
	tree := avltree.CreateAVLTree()

	// Убацујемо елементе
	tree.Insert("apple", 1)
	tree.Insert("apricot", 2)
	tree.Insert("banana", 3)
	tree.Insert("app", 4)
	tree.Insert("application", 5)

	results := tree.PrefixScan("app", 1, 10)

	expected := []int{1, 4, 5} // values за "apple", "app", "application" (редослед зависи од InOrder)
	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}
}

func TestRangeScan(t *testing.T) {
	tree := avltree.CreateAVLTree()

	tree.Insert("a", 1)
	tree.Insert("c", 3)
	tree.Insert("b", 2)
	tree.Insert("e", 5)
	tree.Insert("d", 4)

	results := tree.RangeScan("b", "d", 1, 10)

	expected := []int{2, 3, 4} // вредности за кључеве "b", "c", "d"
	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}
}

func TestDuplicates(t *testing.T) {
	tree := avltree.CreateAVLTree()

	tree.Insert("key", 1)
	tree.Insert("key", 999) // дупликат не би требало да промени постојећи

	results := tree.PrefixScan("key", 1, 10)

	if len(results) != 1 || results[0] != 1 {
		t.Fatalf("Duplicate insert failed: expected [1], got %v", results)
	}
}
