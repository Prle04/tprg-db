package btree

import (
	"fmt"
	"testing"
	"time"
)

func TestBTree_Set(t *testing.T) {
	tree, _ := NewBTree(3)

	tree.Set("key1", []byte("value1"), uint64(time.Now().Unix()))
	tree.Set("key2", []byte("value2"), uint64(time.Now().Unix()))
	tree.Set("key3", []byte("value3"), uint64(time.Now().Unix()))

	if tree.isTreeEmpty() {
		t.Errorf("Tree should not be empty after insertions")
	}

	if el := tree.Get("key1"); string(el.GetValue()) != "value1" {
		t.Errorf("Expected value1, got %s", el.GetValue())
	}

	if el := tree.Get("key2"); string(el.GetValue()) != "value2" {
		t.Errorf("Expected value2, got %s", el.GetValue())
	}

	if el := tree.Get("key3"); string(el.GetValue()) != "value3" {
		t.Errorf("Expected value3, got %s", el.GetValue())
	}
}

func TestBTree_Delete(t *testing.T) {
	tree, _ := NewBTree(3)

	tree.Set("key1", []byte("value1"), uint64(time.Now().Unix()))
	tree.Set("key2", []byte("value2"), uint64(time.Now().Unix()))
	tree.Set("key3", []byte("value3"), uint64(time.Now().Unix()))

	tree.Delete("key2", uint64(time.Now().Unix()))

	if el := tree.Get("key2"); !el.IsDeleted() {
		t.Errorf("Expected to be deleted, got %t", el.IsDeleted())
	}
}

func TestBTree_Get(t *testing.T) {
	tree, _ := NewBTree(3)

	tree.Set("key1", []byte("value1"), uint64(time.Now().Unix()))
	tree.Set("key2", []byte("value2"), uint64(time.Now().Unix()))
	fmt.Println("Set!")
	if el := tree.Get("key1"); string(el.GetValue()) != "value1" {
		t.Errorf("Expected value1, got %s", el.GetValue())
	}
	fmt.Println("Found1!")

	if el := tree.Get("key2"); string(el.GetValue()) != "value2" {
		t.Errorf("Expected value2, got %s", el.GetValue())
	}
	fmt.Println("found2!")

	if el := tree.Get("key3"); el != nil {
		fmt.Println(el)
		t.Errorf("Expected nil, got %s", el.GetKey())
	}
	fmt.Println("Found3!")
}

func TestBTree_SplitNode(t *testing.T) {
	tree, _ := NewBTree(3)

	tree.Set("key1", []byte("value1"), uint64(time.Now().Unix()))
	tree.Set("key2", []byte("value2"), uint64(time.Now().Unix()))
	tree.Set("key3", []byte("value3"), uint64(time.Now().Unix()))
	tree.Set("key4", []byte("value4"), uint64(time.Now().Unix()))

	if tree.root == nil {
		t.Errorf("Root should not be nil")
	}

	if len(tree.root.keys) != 1 {
		t.Errorf("Root should have 1 key after split, got %d", len(tree.root.keys))
	}

	if len(tree.root.children) != 2 {
		t.Errorf("Root should have 2 children after split, got %d", len(tree.root.children))
	}
}
