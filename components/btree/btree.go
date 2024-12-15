package btree

import (
	"fmt"
	"sort"
)

type Element struct {
	key       string
	value     []byte
	isDeleted bool
}

type Node struct {
	//Lista kljuceva
	//Lista dece
	//roditelj
	keys     []*Element
	children []*Node
	parent   *Node
}

type BTree struct {
	//elementi b stabla
	root             *Node
	minChildren      uint32
	maxChildren      uint32
	numberOfElements uint32
}

func (tree *BTree) isTreeEmpty() bool {
	return tree.root == nil
}

func isLeaf(n *Node) bool {
	return len(n.children) == 0
}

func (tree *BTree) splitNode(node *Node) {
	midIndex := len(node.keys) / 2
	midElement := node.keys[midIndex]

	// Kreiramo nove čvorove koji će primiti ključeve i decu sa leve i desne strane srednjeg elementa
	leftNode := &Node{
		keys:     make([]*Element, midIndex),
		children: nil,
		parent:   node.parent,
	}
	rightNode := &Node{
		keys:     make([]*Element, len(node.keys)-midIndex-1),
		children: nil,
		parent:   node.parent,
	}
	copy(leftNode.keys, node.keys[:midIndex])
	copy(rightNode.keys, node.keys[midIndex+1:])

	// Ako čvor ima decu, raspodeljujemo ih na novokreirane čvorove
	if len(node.children) > 0 {
		leftNode.children = make([]*Node, midIndex+1)
		rightNode.children = make([]*Node, len(node.keys)-midIndex)
		copy(leftNode.children, node.children[:midIndex+1])
		copy(rightNode.children, node.children[midIndex+1:])

		for _, child := range leftNode.children {
			child.parent = leftNode
		}
		for _, child := range rightNode.children {
			child.parent = rightNode
		}
	}

	if node.parent == nil {
		// Ako je node root, kreiramo novi root koji sadrži samo srednji element
		tree.root = &Node{
			keys:     []*Element{midElement},
			children: []*Node{leftNode, rightNode},
			parent:   nil,
		}
		leftNode.parent = tree.root
		rightNode.parent = tree.root
	} else {
		// Umetanje srednjeg elementa u roditeljski čvor
		parent := node.parent
		insertIndex := sort.Search(len(parent.keys), func(i int) bool { return midElement.key < parent.keys[i].key })
		parent.keys = append(parent.keys, nil)
		copy(parent.keys[insertIndex+1:], parent.keys[insertIndex:])
		parent.keys[insertIndex] = midElement

		parent.children = append(parent.children, nil)
		copy(parent.children[insertIndex+2:], parent.children[insertIndex+1:])
		parent.children[insertIndex] = leftNode
		parent.children[insertIndex+1] = rightNode

		// Ako se roditeljski čvor prebije, splitujemo i njega
		if uint32(len(parent.keys)) > tree.maxChildren {
			tree.splitNode(parent)
		}
	}
}

func (tree *BTree) findPosition(key string) *Node {
	currentNode := tree.root
	for {
		//Ako je broj dece 0 samo vratimo taj node u kojem ubacujemo element
		if len(currentNode.children) == 0 {
			break
		}

		//Prolazimo kroz kljuceve
		for i, element := range currentNode.keys {
			//Ako je trazeni kljuc manji od trenutnog spustamo se u njegovu decu
			if key < element.key {
				currentNode = currentNode.children[i]
				break
			} else if i == len(currentNode.keys)-1 {
				currentNode = currentNode.children[i+1]
			}
		}
	}

	return currentNode
}

func (tree *BTree) findNode(key string) (*Node, *Element) {
	if tree.isTreeEmpty() {
		return nil, nil
	}

	//Polazimo od roota
	currentNode := tree.root

	//Prolazimo kroz cvorove (Node)
	for currentNode != nil {

		//Prolazimo kroz svaki kljuc u tom cvoru
		for i, element := range currentNode.keys {
			//Ako smo pronasli key vracamo trazeni node
			if key == element.key {
				return currentNode, element
				//Ako je trazeni key manji od trenutnog znaci potencijalno se nalazi kod dece
			} else if key < element.key && len(currentNode.children) != 0 {
				currentNode = currentNode.children[i]
				break
				//Ako nije nigde onda silazimo na sledece dete
			} else if i == len(currentNode.keys)-1 && len(currentNode.children) != 0 {
				currentNode = currentNode.children[i+1]
			} else if i == len(currentNode.keys)-1 && len(currentNode.children) == 0 {
				return nil, nil
			}
		}

	}
	return nil, nil
}

func (tree *BTree) insertElement(elem *Element) {
	if tree.isTreeEmpty() {
		tree.root = &Node{
			keys:     []*Element{elem},
			children: []*Node{},
			parent:   nil,
		}
		tree.numberOfElements++
		return
	}

	_, existingElement := tree.findNode(elem.key)

	if existingElement != nil {
		existingElement.value = elem.value
		existingElement.isDeleted = elem.isDeleted
		return
	}

	currentNode := tree.findPosition(elem.key)
	insertIndex := sort.Search(len(currentNode.keys), func(i int) bool { return elem.key <= currentNode.keys[i].key })
	currentNode.keys = append(currentNode.keys, nil)
	copy(currentNode.keys[insertIndex+1:], currentNode.keys[insertIndex:])
	currentNode.keys[insertIndex] = elem
	tree.numberOfElements += 1

	if uint32(len(currentNode.keys)) > tree.maxChildren {
		tree.splitNode(currentNode)
	}

}

func (tree *BTree) printNode(node *Node, level int) {
	if node == nil {
		return
	}

	// Print the current node's keys
	fmt.Printf("Level %d: ", level)
	for _, element := range node.keys {
		if element.isDeleted {
			fmt.Printf("(%s, !) ", element.key)
		} else {
			fmt.Printf("(%s) ", element.key)
		}
	}
	fmt.Println()

	// Recursively print the children nodes
	for _, child := range node.children {
		tree.printNode(child, level+1)
	}
}

func (tree *BTree) Set(key string, value []byte) {
	//Kreiramo novi element
	newElement := &Element{
		key:       key,
		value:     value,
		isDeleted: false,
	}

	tree.insertElement(newElement)
}

func (tree *BTree) Delete(key string) {
	currentNode, elem := tree.findNode(key)

	if currentNode == nil {
		tree.insertElement(&Element{key: key, value: nil, isDeleted: true})
	} else {
		elem.isDeleted = true
	}
}

func (tree *BTree) Get(key string) []byte {
	if tree.isTreeEmpty() {
		return nil
	}

	currentNode, elem := tree.findNode(key)
	if currentNode == nil {
		return nil
	}

	if !elem.isDeleted {
		return elem.value
	}

	return nil
}

func (tree *BTree) Print() {
	if tree.isTreeEmpty() {
		fmt.Println("The BTree is empty")
		return
	}
	tree.printNode(tree.root, 0)
}

func NewBTree(maxChildren uint32) (*BTree, error) {
	if maxChildren <= 0 {
		return nil, &MaxChildrenError{}
	}
	return &BTree{
		root:             nil,
		minChildren:      maxChildren / 2,
		maxChildren:      maxChildren,
		numberOfElements: 0,
	}, nil
}

func (tree *BTree) GetNumberOfKeys() uint32 {
	return tree.numberOfElements
}

type MaxChildrenError struct{}

func (m *MaxChildrenError) Error() string {
	return "number of children must be positive integer!"
}
