package merkletree_test

import (
	"testing"

	"tprgdb/components/merkletree"
)

func TestNewMerkleTree_NilInput(t *testing.T) {
	tree := merkletree.NewMerkleTree(nil)
	if tree != nil {
		t.Errorf("Expected nil, got %v", tree)
	}
}

func TestNewMerkleTree_EmptyInput(t *testing.T) {
	tree := merkletree.NewMerkleTree([]*[]byte{})
	if tree != nil {
		t.Errorf("Expected nil, got %v", tree)
	}
}


func TestNewMerkleTree_ValidInput(t *testing.T) {
	D1 := []byte("data1")
	D2 := []byte("data2")
	D3 := []byte("data3")
	D4 := []byte("data4")
	D5 := []byte("data5")
	data := []*[]byte{&D1, &D2, &D3, &D4, &D5}
	tree := merkletree.NewMerkleTree(data)
	if tree == nil {
		t.Error("Expected non-nil tree, got nil")
	}
}

func TestMerkleTree_GetRoot(t *testing.T) {
	D1 := []byte("data1")
	D2 := []byte("data2")
	D3 := []byte("data3")
	D4 := []byte("data4")
	data := []*[]byte{&D1, &D2, &D3, &D4}
	tree := merkletree.NewMerkleTree(data)
	if tree == nil {
		t.Fatal("Expected non-nil tree, got nil")
	}

	root := tree.GetRoot()
	if root == 0 {
		t.Error("Expected non-zero root hash, got 0")
	}
}

func TestMerkleTree_SameRoot(t *testing.T) {D1 := []byte("data1")
	D2 := []byte("data2")
	D3 := []byte("data3")
	D4 := []byte("data4")
	D5 := []byte("data5")
	data := []*[]byte{&D1, &D2, &D3, &D4, &D5}
	tree := merkletree.NewMerkleTree(data)
	if tree == nil {
		t.Error("Expected non-nil tree, got nil")
	}
	data2 := []*[]byte{&D1, &D2, &D3, &D4, &D5}
	tree2 := merkletree.NewMerkleTree(data2)
	if tree2 == nil {
		t.Error("Expected non-nil tree, got nil")
	}
	if tree.GetRoot() != tree2.GetRoot() {
		t.Errorf("Expected roots to be equal, got %v and %v", tree.GetRoot(), tree2.GetRoot())
	}
}

func TestMerkleTree_DifferentData(t *testing.T) {
	D1 := []byte("data1")
	D2 := []byte("data2")
	D3 := []byte("data3")
	D4 := []byte("data4")
	data := []*[]byte{&D1, &D2, &D3, &D4}
	tree := merkletree.NewMerkleTree(data)
	if tree == nil {
		t.Error("Expected non-nil tree, got nil")
	}

	D5 := []byte("data5")
	data2 := []*[]byte{&D1, &D2, &D3, &D4, &D5}
	tree2 := merkletree.NewMerkleTree(data2)
	if tree2 == nil {
		t.Error("Expected non-nil tree, got nil")
	}

	if tree.GetRoot() == tree2.GetRoot() {
		t.Errorf("Expected different roots for different data, got %v and %v", tree.GetRoot(), tree2.GetRoot())
	}
}
