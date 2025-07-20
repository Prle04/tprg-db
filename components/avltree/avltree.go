package avltree

import (
	"bytes"
	"fmt"
	"strings"
	"tprgdb/components/blockmanager"
	"tprgdb/components/serialization"
)

const CURRENT_VERSION uint8 = 1

type ISerializable interface {
	SerializeToBytes() []byte
	DeserializeFromBytes(data []byte)
	SerializeToFile(file *blockmanager.BlockedFile) error
	DeserializeFromFile(file *blockmanager.BlockedFile) error
}

type Node struct {
	key        string
	value      uint64
	height     int
	leftChild  *Node
	rightChild *Node
}

type Data struct {
	key   string
	value uint64
}

func (d *Data) GetKey() string {
	return d.key
}

func (d *Data) GetValue() uint64 {
	return d.value
}

func createData(key string, value uint64) *Data {
	return &Data{
		key:   key,
		value: value,
	}
}

type AVLTree struct {
	root             *Node
	numberOfElements uint64
}

func CreateAVLTree() *AVLTree {
	return &AVLTree{
		root:             nil,
		numberOfElements: 0,
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
	x := y.leftChild
	T2 := x.rightChild

	x.rightChild = y
	y.leftChild = T2

	updateHeight(y)
	updateHeight(x)
	return x
}

func rotateLeft(x *Node) *Node {
	y := x.rightChild
	T2 := y.leftChild

	y.leftChild = x
	x.rightChild = T2

	updateHeight(x)
	updateHeight(y)
	return y
}

func (tree *AVLTree) insert(node *Node, key string, value uint64) *Node {
	if node == nil {
		tree.numberOfElements++
		return &Node{key: key, height: 1, value: value}
	}

	if key < node.key {
		node.leftChild = tree.insert(node.leftChild, key, value)
	} else if key > node.key {
		node.rightChild = tree.insert(node.rightChild, key, value)
	} else {
		return node
	}

	updateHeight(node)
	balance := getBalance(node)

	if balance > 1 && key < node.leftChild.key {
		return rotateRight(node)
	}

	if balance < -1 && key > node.rightChild.key {
		return rotateLeft(node)
	}

	if balance > 1 && key > node.leftChild.key {
		node.leftChild = rotateLeft(node.leftChild)
		return rotateRight(node)
	}

	if balance < -1 && key < node.rightChild.key {
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
		return find(node.leftChild, key)
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

func (tree *AVLTree) Insert(key string, value uint64) {
	tree.root = tree.insert(tree.root, key, value)
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

func (tree *AVLTree) PrefixScan(prefix string, page int, perPage int) []*Data {
	var result []*Data
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
		if len(result) == perPage {
			return
		}
		if n.key >= lowerBound && n.key <= upperBound {
			if strings.HasPrefix(n.key, prefix) {
				if count >= startIndex {
					newData := createData(n.key, n.value)
					result = append(result, newData)
				}
				count++
				if len(result) == perPage {
					return
				}
			}
		}
		if n.key <= upperBound {
			helper(n.rightChild)
		}
	}
	helper(tree.root)
	return result
}

func (tree *AVLTree) RangeScan(from, to string, page, perPage int) []*Data {
	var result []*Data
	var scan func(n *Node)

	startIndex := (page - 1) * perPage
	count := 0

	scan = func(n *Node) {
		if n == nil {
			return
		}
		if n.key >= from {
			scan(n.leftChild)
		}
		if len(result) == perPage {
			return
		}
		if n.key >= from && n.key <= to {
			if count >= startIndex {
				newData := createData(n.key, n.value)
				result = append(result, newData)
			}
			count++
			if len(result) == perPage {
				return
			}
		}
		if n.key <= to {
			scan(n.rightChild)
		}
	}
	scan(tree.root)
	return result
}

func (tree *AVLTree) Find(key string) uint64 {
	node := find(tree.root, key)
	if node == nil {
		return 0
	}
	return node.value
}

func (tree *AVLTree) GetNumberOfElements() uint64 {
	return tree.numberOfElements
}

func (tree *AVLTree) SerializeToBytes() []byte {
	var buf bytes.Buffer

	buf.Write(serialization.UInt8ToBytes(CURRENT_VERSION))

	buf.Write(serialization.UInt64ToBytes(tree.numberOfElements))

	var serializeNode func(n *Node)
	serializeNode = func(n *Node) {
		if n == nil {
			buf.Write(serialization.Int8ToBytes(0))
			return
		}
		buf.Write(serialization.Int8ToBytes(1))

		keyBytes := serialization.StringToBytes(n.key)
		buf.Write(serialization.UInt64ToBytes(uint64(len(keyBytes))))
		buf.Write(keyBytes)

		buf.Write(serialization.UInt64ToBytes(n.value))

		serializeNode(n.leftChild)
		serializeNode(n.rightChild)
	}
	serializeNode(tree.root)

	return buf.Bytes()
}

func (tree *AVLTree) DeserializeFromBytes(data []byte) {
	buf := bytes.NewReader(data)

	versionBuf := make([]byte, 1)
	buf.Read(versionBuf)
	version := serialization.BytesToUInt8(versionBuf)
	if version != CURRENT_VERSION {
		panic("unsupported version")
	}

	numBuf := make([]byte, 8)
	buf.Read(numBuf)
	tree.numberOfElements = serialization.BytesToUInt64(numBuf)

	var deserializeNode func() *Node
	deserializeNode = func() *Node {
		markerBuf := make([]byte, 1)
		buf.Read(markerBuf)
		marker := serialization.BytesToInt8(markerBuf)
		if marker == 0 {
			return nil
		}

		lenBuf := make([]byte, 8)
		buf.Read(lenBuf)
		keyLen := serialization.BytesToUInt64(lenBuf)

		keyBytes := make([]byte, keyLen)
		buf.Read(keyBytes)
		key := serialization.BytesToString(keyBytes)

		valBuf := make([]byte, 8)
		buf.Read(valBuf)
		value := serialization.BytesToUInt64(valBuf)

		node := &Node{
			key:    key,
			value:  value,
			height: 1,
		}
		node.leftChild = deserializeNode()
		node.rightChild = deserializeNode()
		updateHeight(node)
		return node
	}
	tree.root = deserializeNode()
}

func (tree *AVLTree) SerializeToFile(file *blockmanager.BlockedFile) error {
	data := tree.SerializeToBytes()
	return file.Write(data)
}

func (tree *AVLTree) DeserializeFromFile(file *blockmanager.BlockedFile) error {
	versionBytes, err := file.Read(1)
	if err != nil {
		return err
	}
	version := serialization.BytesToUInt8(versionBytes)
	if version != CURRENT_VERSION {
		return fmt.Errorf("unsupported version: %d", version)
	}

	numBytes, err := file.Read(8)
	if err != nil {
		return err
	}
	tree.numberOfElements = serialization.BytesToUInt64(numBytes)

	var deserializeNode func() (*Node, error)
	deserializeNode = func() (*Node, error) {
		markerBytes, err := file.Read(1)
		if err != nil {
			return nil, err
		}
		marker := serialization.BytesToInt8(markerBytes)
		if marker == 0 {
			return nil, nil
		}

		keyLenBytes, err := file.Read(8)
		if err != nil {
			return nil, err
		}
		keyLen := serialization.BytesToUInt64(keyLenBytes)

		keyBytes, err := file.Read(keyLen)
		if err != nil {
			return nil, err
		}
		key := serialization.BytesToString(keyBytes)

		valBytes, err := file.Read(8)
		if err != nil {
			return nil, err
		}
		value := serialization.BytesToUInt64(valBytes)

		node := &Node{
			key:    key,
			value:  value,
			height: 1,
		}

		node.leftChild, err = deserializeNode()
		if err != nil {
			return nil, err
		}
		node.rightChild, err = deserializeNode()
		if err != nil {
			return nil, err
		}

		updateHeight(node)
		return node, nil
	}

	root, err := deserializeNode()
	if err != nil {
		return err
	}
	tree.root = root
	return nil
}
