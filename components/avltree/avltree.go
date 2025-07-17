package avltree

import (
	"fmt"
	"strings"
	"tprgdb/components/blockmanager"
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

// ----------------------------- AVL LOGIKA -----------------------------
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

func (tree *AVLTree) Find(key string) *Data {
	node := find(tree.root, key)
	return createData(node.key, node.value)
}

func (tree *AVLTree) GetNumberOfElements() uint64 {
	return tree.numberOfElements
}

// ------------------ S E R I J A L I Z A C И Ј А ------------------
/*
func (tree *AVLTree) SerializeToBytes() []byte {
	var buf bytes.Buffer
	buf.WriteByte(CURRENT_VERSION)

	// broj elemenata
	buf.Write(serialization.UVarInt(tree.numberOfElements))

	// rekurzivno čvorovi
	var serializeNode func(n *Node)
	serializeNode = func(n *Node) {
		if n == nil {
			buf.WriteByte(0) // NIL marker
			return
		}
		buf.WriteByte(1) // postoji čvor

		keyBytes := []byte(n.key)
		buf.Write(serialization.UVarInt(uint64(len(keyBytes))))
		buf.Write(keyBytes)

		buf.Write(serialization.UVarInt(n.value))
		buf.Write(serialization.UVarInt(uint64(n.height)))

		serializeNode(n.leftChild)
		serializeNode(n.rightChild)
	}

	serializeNode(tree.root)
	return buf.Bytes()
}

func (tree *AVLTree) DeserializeFromBytes(data []byte) {
	if len(data) == 0 || data[0] != CURRENT_VERSION {
		return
	}
	data = data[1:]

	var idx int

	readUVar := func() uint64 {
		val, n := binary.Uvarint(data[idx:])
		idx += n
		return val
	}
	readBytes := func(length uint64) []byte {
		res := data[idx : idx+int(length)]
		idx += int(length)
		return res
	}

	tree.numberOfElements = readUVar()

	var deserializeNode func() *Node
	deserializeNode = func() *Node {
		if idx >= len(data) {
			return nil
		}
		if data[idx] == 0 {
			idx++
			return nil
		}
		idx++ // skip marker

		keyLen := readUVar()
		key := string(readBytes(keyLen))
		val := readUVar()
		height := int(readUVar())

		node := &Node{
			key:    key,
			value:  val,
			height: height,
		}
		node.leftChild = deserializeNode()
		node.rightChild = deserializeNode()
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
	if versionBytes[0] != CURRENT_VERSION {
		return errors.New("version mismatch")
	}

	tree.numberOfElements = 0

	var deserializeNode func() (*Node, error)
	deserializeNode = func() (*Node, error) {
		marker, err := file.Read(1)
		if err != nil {
			return nil, err
		}
		if marker[0] == 0 {
			return nil, nil
		}

		_, keyLen, err := file.ReadUVariable()
		if err != nil {
			return nil, err
		}
		keyBytes, err := file.Read(keyLen)
		if err != nil {
			return nil, err
		}
		key := string(keyBytes)

		_, val, err := file.ReadUVariable()
		if err != nil {
			return nil, err
		}

		_, h, err := file.ReadUVariable()
		if err != nil {
			return nil, err
		}

		node := &Node{
			key:    key,
			value:  val,
			height: int(h),
		}

		node.leftChild, err = deserializeNode()
		if err != nil {
			return nil, err
		}
		node.rightChild, err = deserializeNode()
		if err != nil {
			return nil, err
		}

		tree.numberOfElements++
		return node, nil
	}

	root, err := deserializeNode()
	if err != nil {
		return err
	}
	tree.root = root
	return nil
}
*/
