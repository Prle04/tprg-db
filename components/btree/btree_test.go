package btree

import (
	"fmt"
	"testing"
)

func TestBTree_Push(t *testing.T) {
	tree := NewBTree(3)

	tree.Push("key1", []byte("value1"))
	tree.Push("key2", []byte("value2"))
	tree.Push("key3", []byte("value3"))

	if tree.isTreeEmpty() {
		t.Errorf("Tree should not be empty after insertions")
	}

	if value := tree.Find("key1"); string(value) != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	if value := tree.Find("key2"); string(value) != "value2" {
		t.Errorf("Expected value2, got %s", value)
	}

	if value := tree.Find("key3"); string(value) != "value3" {
		t.Errorf("Expected value3, got %s", value)
	}
}

func TestBTree_Pop(t *testing.T) {
	tree := NewBTree(3)

	tree.Push("key1", []byte("value1"))
	tree.Push("key2", []byte("value2"))
	tree.Push("key3", []byte("value3"))

	tree.Pop("key2")

	if value := tree.Find("key2"); value != nil {
		t.Errorf("Expected nil, got %s", value)
	}
}

func TestBTree_Find(t *testing.T) {
	tree := NewBTree(3)

	tree.Push("key1", []byte("value1"))
	tree.Push("key2", []byte("value2"))
	fmt.Println("Pushed!")
	if value := tree.Find("key1"); string(value) != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}
	fmt.Println("Found1!")

	if value := tree.Find("key2"); string(value) != "value2" {
		t.Errorf("Expected value2, got %s", value)
	}
	fmt.Println("found2!")

	if value := tree.Find("key3"); value != nil {
		fmt.Println(value)
		t.Errorf("Expected nil, got %s", value)
	}
	fmt.Println("Found3!")
}

func TestBTree_SplitNode(t *testing.T) {
	tree := NewBTree(3)

	tree.Push("key1", []byte("value1"))
	tree.Push("key2", []byte("value2"))
	tree.Push("key3", []byte("value3"))
	tree.Push("key4", []byte("value4"))

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
