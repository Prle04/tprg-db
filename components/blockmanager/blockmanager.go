package blockmanager

import (
	"os"
	"strconv"
	"strings"
	"tprgdb/components/lrucache"
)

type BlockManager struct {
	blockSize   uint64
	maxSize     uint64
	currentSize uint64
	blocks      map[string]*Block
	cache       *lrucache.LRUCache[*Block]
}


func NewBlockManager(blockSize uint64, maxSize uint64) *BlockManager {
	return &BlockManager{
		blockSize:   blockSize,
		maxSize:     maxSize,
		currentSize: 0,
		blocks:      make(map[string]*Block),
		cache:       lrucache.NewLRUCache[*Block](maxSize),
	}
}

func (m *BlockManager) OpenBlockedFile(path string) *BlockedFile {
	return &BlockedFile{
		manager:  m,
		path:     path,
		position: 0,
		blockSize: m.blockSize,
	}
}

func (m *BlockManager) getBlock(path string) (*Block, error) {
	// Block is already in the cache
	if block, ok := m.blocks[path]; ok {
		return block, nil
	}
	// Block is not cached so we load it from the file
	return m.load(path)
}

func (m *BlockManager) addBlock(block *Block) error {
	if _, ok := m.blocks[block.GetPath()]; !ok {
		m.currentSize++
	}
	m.blocks[block.GetPath()] = block
	evicted, evictedPath := m.cache.AddElement(block.GetPath(), block)
	if evicted {
		err := m.syncBlock(evictedPath)
		if err != nil {
			return err
		}
		m.removeBlock(evictedPath)
	}
	return nil
}

func (m *BlockManager) removeBlock(path string) {
	if _, ok := m.blocks[path]; !ok {
		return
	}
	m.cache.RemoveElement(path)
	delete(m.blocks, path)
	m.currentSize--
}

func (m *BlockManager) syncBlock(path string) error {
	block, ok := m.blocks[path]
	if !ok {
		return nil
	}
	
	if !block.IsModified() {
		return nil
	}
	file, err := os.OpenFile(block.filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Seek(int64(m.blockSize*block.orderNumber), 0)
	if err != nil {
		return err
	}
	_, err = file.Write(block.data)
	if err != nil {
		return err
	}
	return nil
}

func (m *BlockManager) syncFile(filename string) error {
	blocks := make([]*Block, 0)

	for _, block := range m.blocks {
		if block.filename == filename {
			blocks = append(blocks, block)
		}
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, block := range blocks {
		_, err = file.Seek(int64(m.blockSize*block.orderNumber), 0)
		if err != nil {
			return err
		}
		_, err = file.Write(block.data)
		if err != nil {
			return err
		}
	}
	return nil
}


func (m *BlockManager) closeFile(filepath string) error {
	syncErr := m.syncFile(filepath)
	if syncErr != nil {
		return syncErr
	}
	for blockpath, block := range m.blocks {
		if block.filename == filepath {
			m.removeBlock(blockpath)
		}
	}
	return nil
}

func (m *BlockManager) deleteFile(filepath string) error {
	for _, block := range m.blocks {
		if block.filename == filepath {
			m.removeBlock(block.filename)
		}
	}
	return os.Remove(filepath)
}

func (m *BlockManager) load(path string) (*Block, error) {
	// Parse the path to get the file path and order number
	var filepath string = getBlockFilename(path)
	orderNumber, err := getBlockOrderNumber(path)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	// Check if the block is in the file
	if fileInfo.Size() <= int64(m.blockSize*orderNumber) {
		// The block is not in the file, so we create a new block
		block := NewBlock(m.blockSize, filepath)
		block.SetOrderNumber(orderNumber)
		err = m.addBlock(block)
		return block, err
	}
	// Create a new block and read the data from the file
	block := NewBlock(m.blockSize, filepath)
	block.SetOrderNumber(orderNumber)
	_, err = file.Seek(int64(m.blockSize*orderNumber), 0)
	if err != nil {
		return nil, err
	}
	_, err = file.Read(block.data)
	if err != nil {
		return nil, err
	}
	// Add the block to the cache
	err = m.addBlock(block)
	if err != nil {
		return nil, err
	}
	return block, nil
}


func getBlockPath(filename string, orderNumber uint64) string {
	return filename + "+" + strconv.FormatUint(orderNumber, 10)
}

func getBlockOrderNumber(path string) (uint64, error) {
	split := strings.Split(path, "+")
	if len(split) != 2 {
		return 0, &InvalidBlockPath{}
	}
	orderNumber, err := strconv.ParseUint(split[len(split)-1], 10, 64)
	if err != nil {
		return 0, err
	}
	return orderNumber, nil
}

func getBlockFilename(path string) string {
	split := strings.Split(path, "+")
	return strings.Join(split[:len(split)-1], "+")
}