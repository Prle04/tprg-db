package avltree

import (
	"fmt"
	"strings"
)

type Node struct {
	key        string
	value      int
	height     int
	leftChild  *Node
	rightChild *Node
}




// Dodaj broj elemenata
type AVLTree struct {
	root *Node
}

func CreateAVLTree() *AVLTree {
	return &AVLTree{
		root: nil,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func height(n *Node) int {
	if n == nil {
		return 0
	}
	return n.height
}

func getBalance(n *Node) int {
	if n == nil {
		return 0
	}
	return height(n.leftChild) - height(n.rightChild)
}

func updateHeight(n *Node) {
	n.height = 1 + max(height(n.leftChild), height(n.rightChild))
}

func rotateRight(y *Node) *Node {
	//Ovde rotiram udesno
	x := y.leftChild
	T2 := x.rightChild

	x.rightChild = y
	y.leftChild = T2

	updateHeight(y)
	updateHeight(x)
	return x
}

func rotateLeft(x *Node) *Node {
	//Rotira udesno
	y := x.rightChild
	T2 := y.leftChild

	y.leftChild = x
	x.rightChild = T2

	updateHeight(x)
	updateHeight(y)
	return y
}

func insert(node *Node, key string, value int) *Node {
	if node == nil {
		//Insertuje se novi node
		return &Node{key: key, height: 1, value: value}
	}

	if key < node.key {
		node.leftChild = insert(node.leftChild, key, value)
	} else if key > node.key {
		node.rightChild = insert(node.rightChild, key, value)
	} else {
		//Ako kljuc vec postoji samo vracamo trenutni node
		return node
	}

	updateHeight(node)
	balance := getBalance(node)

	//Proveravamo da li je narusen balans

	if balance > 1 && key < node.leftChild.key {
		//Rotiramo desno
		return rotateRight(node)
	}

	if balance < -1 && key > node.rightChild.key {
		//rotiramo levo
		return rotateLeft(node)
	}

	if balance > 1 && key > node.leftChild.key {
		//Rotiramo levo pa desno
		node.leftChild = rotateLeft(node.leftChild)
		return rotateRight(node)
	}

	if balance < -1 && key < node.rightChild.key {
		//Rotiramo desno pa levo
		node.rightChild = rotateRight(node.rightChild)
		return rotateLeft(node)
	}

	return node
}

func find(node *Node, key string) *Node {
	if node == nil {
		return nil
	}
	if key < node.key {
		return find(node.rightChild.leftChild, key)
	} else if key > node.key {
		return find(node.rightChild, key)
	}
	return node
}

func inOrder(n *Node) {
	if n != nil {
		inOrder(n.leftChild)
		fmt.Print(n.key, " ")
		inOrder(n.rightChild)
	}
}

func preOrder(n *Node) {
	if n != nil {
		fmt.Print(n.key, " ")
		preOrder(n.leftChild)
		preOrder(n.rightChild)
	}
}

func postOrder(n *Node) {
	if n != nil {
		postOrder(n.leftChild)
		postOrder(n.rightChild)
		fmt.Print(n.key, " ")
	}
}

func (tree *AVLTree) Insert(key string, value int) {
	tree.root = insert(tree.root, key, value)
}

func (tree *AVLTree) InOrder() {
	inOrder(tree.root)
	fmt.Println()
}

func (tree *AVLTree) PreOrder() {
	preOrder(tree.root)
	fmt.Println()
}

func (tree *AVLTree) PostOrder() {
	postOrder(tree.root)
	fmt.Println()
}

// Ne vracaj []int već []{key, value}
func (tree *AVLTree) PrefixScan(prefix string, page int, perPage int) []int {
	var result []int
	lowerBound := prefix
	upperBound := prefix + "\uffff"

	startIndex := (page - 1) * perPage
	count := 0

	var helper func(n *Node)
	helper = func(n *Node) {
		if n == nil {
			return
		}

		if n.key >= lowerBound {
			helper(n.leftChild)
		}

		if n.key >= lowerBound && n.key <= upperBound {
			if strings.HasPrefix(n.key, prefix) {
				if count >= startIndex {
					result = append(result, n.value)
				}
				if len(result) == perPage {
					return
				}
				count++
			}
		}

		if n.key <= upperBound {
			helper(n.rightChild)
		}
	}

	helper(tree.root)
	return result
}

// Ne vracaj []int već []{key, value}
func (tree *AVLTree) RangeScan(from, to string, page, perPage int) []int {
	var result []int
	var scan func(n *Node)

	startIndex := (page - 1) * perPage
	count := 0

	scan = func(n *Node) {
		if n == nil {
			return
		}

		if n.key > from {
			scan(n.leftChild)
		}
		if n.key >= from && n.key <= to {
			if count >= startIndex {
				result = append(result, n.value)
			}
			if len(result) == perPage {
				return
			}
			count++
		}
		if n.key < to {
			scan(n.rightChild)
		}
	}

	scan(tree.root)
	return result
}

//Ne treba mi kurcu nego mi bilo zabano da pisem
func (tree *AVLTree) Find(key string) *Node {
	return find(tree.root,key)
}
