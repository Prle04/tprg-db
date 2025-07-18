package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/lsm"
	"tprgdb/components/sstable"
)

var testDir string = "test-temp"

func newTestLSM(maxLevel int, algo string, filePerLevel int, multiplier int) *lsm.LSM {
	blockMgr := blockmanager.NewBlockManager(2048, 128)
	compression := NewCompressionMockup(true)
	sstMgr := sstable.NewSSTableManager(testDir, "one-file", 2, blockMgr, compression)

	tree := lsm.NewLSM(maxLevel, testDir, algo, blockMgr, sstMgr, compression, filePerLevel, multiplier)
	if tree == nil {
		panic("Failed to create LSM tree")
	}
	return tree
}

var BM = blockmanager.NewBlockManager(2048, 128)

// func setupTestDir() {
// 	testDir = filepath.Join("test-temp")
// 	err := os.MkdirAll(testDir, 0o755)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// }

// func cleanupTestDir() {
// 	if testDir != "" {
// 		err := os.RemoveAll(testDir)
// 		if err != nil {
// 			panic("failed to remove test dir: " + err.Error())
// 		}
// 		testDir = ""
// 	}
// }

type testItem struct {
	key           string
	compressedKey uint64
	value         []byte
}

func (item testItem) GetKey() string {
	return item.key
	// return randomString(10)
}
func (item testItem) GetCompressedKey() uint64 {
	// Mock compression: just return a random uint64
	return item.compressedKey
}

func (item testItem) GetValue() []byte {
	return item.value
}

func (item testItem) GetTimestamp() uint64 {
	return uint64(time.Now().Unix())
}

func (item testItem) IsDeleted() bool {
	return rand.Intn(2) == 0 // Randomly return true or false
}

// func randomString(n int) string {
// 	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
// 	b := make([]rune, n)
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 	}
// 	return string(b)
// }

func getData(n int, comcompressionManager sstable.ICompressionManager) []*testItem {
	items := make([]*testItem, n)
	for i := 0; i < n; i++ {
		items[i] = &testItem{
			key:   fmt.Sprintf("key-%d", i),
			value: []byte(fmt.Sprintf("value-%d", i)),
		}
		items[i].compressedKey, _ = comcompressionManager.GetCompressed(items[i].key)
	}
	return items
}

type compressionMockup struct {
	s2i        map[string]uint64
	i2s        map[uint64]string
	count      uint64
	compressed bool
}

func NewCompressionMockup(compressed bool) *compressionMockup {
	return &compressionMockup{
		s2i:        make(map[string]uint64),
		i2s:        make(map[uint64]string),
		count:      0,
		compressed: compressed,
	}
}

func (c *compressionMockup) GetCompressed(s string) (uint64, error) {
	if !c.compressed {
		return 0, nil // No compression
	}
	if compressed, ok := c.s2i[s]; ok {
		return compressed, nil
	}
	c.count++
	c.s2i[s] = c.count
	c.i2s[c.count] = s
	return c.count, nil
}

func (c *compressionMockup) GetDecompressed(u uint64) (string, error) {
	if !c.compressed {
		return "", nil // No compression
	}
	if decompressed, ok := c.i2s[u]; ok {
		return decompressed, nil
	}
	return "", fmt.Errorf("compressed value %d not found", u)
}

func (c *compressionMockup) IsCompressionEnabled() bool {
	return c.compressed
}

func splitWithOverlap(items []sstable.ISSTableItem, n int, itemsPerSlice int) [][]sstable.ISSTableItem {
	N := len(items)
	toAdd := n*itemsPerSlice - N

	// Append random items (with overlap allowed)
	for range toAdd {
		items = append(items, items[rand.Intn(N)])
	}

	// Shuffle all items
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	// Initialize slices and sets to track unique items per slice
	slices := make([][]sstable.ISSTableItem, n)
	seen := make([]map[string]bool, n) // assuming item.Key() returns string
	for i := range seen {
		seen[i] = make(map[string]bool)
	}

	// Distribute items
	for _, item := range items {
		key := item.GetKey() // adjust this if you use something else as identity

		// Try to place item in a slice that doesn't already contain it
		for j := range n {
			idx := j
			if len(slices[idx]) < itemsPerSlice && !seen[idx][key] {
				slices[idx] = append(slices[idx], item)
				seen[idx][key] = true
				break
			}
		}
	}

	return slices
}

func merge() {

	compressionManager := NewCompressionMockup(true)

	sstableManger := sstable.NewSSTableManager(testDir, sstable.MODE_ONE_FILE, 10, BM, compressionManager)

	data := getData(20, compressionManager)
	items := make([]sstable.ISSTableItem, len(data))
	keys := make([]string, len(data))
	for i, v := range data {
		items[i] = v
		keys[i] = v.GetKey()
	}

	slices := splitWithOverlap(items, 5, 10)

	readers := make([]*sstable.SSTableReader, len(slices))
	for i, slice := range slices {
		sort.Slice(slice, func(i, j int) bool {
			if !compressionManager.IsCompressionEnabled() {
				return slice[i].GetKey() < slice[j].GetKey()
			}
			return slice[i].GetCompressedKey() < slice[j].GetCompressedKey()
		})
		if len(slice) == 0 {
			panic("SLICE IS EMPTYYYYYYYYYYYYY!!!!!!!!!!!!!")
		}

		tableName := fmt.Sprintf("test-table-merge-%d", i)
		reader, err := sstableManger.CreateSSTable(tableName, sstable.NewSSListIterator(slice, compressionManager), uint64(len(slice)))
		if err != nil {
			panic(fmt.Sprintf("failed to create SSTable %s: %v", tableName, err))
		}
		if reader == nil {
			panic(fmt.Sprintf("failed to create SSTable reader for %s", tableName))
		}
		readers[i] = reader
		iter := reader.GetIterator()
		f, err := os.OpenFile(fmt.Sprintf("test-iter-%d.txt", i), os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			panic(fmt.Sprintf("failed to open file for iterator %d: %v", i, err))
		}
		for item, err := iter.Next(); ; item, err = iter.Next() {
			if err == blockmanager.EOF {
				f.Close()
				break
			}
			if err != nil {
				panic(fmt.Sprintf("iterator error: %v", err))
			}
			key := item.GetKey()
			f.WriteString(key + "\n")
		}
	}
	// return
	iter, err := sstable.NewSSTableMergeIterator(readers, compressionManager.IsCompressionEnabled())
	if err != nil {
		panic(fmt.Sprintf("failed to create SSTable merge iterator: %v", err))
	}
	file, _ := os.OpenFile("test-merge.txt", os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	// Write the merged items to a file
	for item, err := iter.Next(); ; item, err = iter.Next() {
		if err == blockmanager.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("merge iterator error: %v", err))
		}
		key := item.GetKey()
		file.WriteString(key + "\n")
	}
}

func LSMTest() {
	tree := newTestLSM(3, lsm.SIZE_TIERED, 2, 10)

	item := &testItem{
		key:           "foo",
		value:         []byte("bar"),
		compressedKey: 42,
	}

	if err := tree.AddTable([]sstable.ISSTableItem{item}); err != nil {
		panic("AddTable failed: %v")
	}

	err := tree.Save()
	if err != nil {
		panic("SaveTable failed: %v")
	}

	// Create new tree, load from disk
	newTree := newTestLSM(3, lsm.SIZE_TIERED, 2, 10)

	err = newTree.Load()
	if err != nil {
		panic("LoadTable failed: %v")
	}

	got, err := tree.Get("foo")
	if err != nil {
		panic("Get failed: %v")
	}
	if got == nil {
		panic("Get failed: nil")
	}
	if string(got.GetValue()) != "bar" {
		panic("Get failed: expected 'bar', got " + string(got.GetValue()))
	}
	fmt.Println("OK")

}
func main() {

	LSMTest()
	var input string
	fmt.Print("Enter a key to search (or 'q' to quit): ")
	_, err := fmt.Scanln(&input)
	if err != nil {
		fmt.Println("Error reading input:", err)
	}

	compressionManager := NewCompressionMockup(false)

	sstableManger := sstable.NewSSTableManager(testDir, sstable.MODE_ONE_FILE, 10, BM, compressionManager)

	data := getData(100, compressionManager)
	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.ISSTableItem, len(data))
	for i, v := range data {
		items[i] = v
	}

	// Sort items by key
	sort.Slice(items, func(i, j int) bool {
		if !compressionManager.IsCompressionEnabled() {
			return items[i].GetKey() < items[j].GetKey()
		}
		return items[i].GetCompressedKey() < items[j].GetCompressedKey()
	})

	iterator1 := sstable.NewSSListIterator(items[:50], compressionManager)
	iterator2 := sstable.NewSSListIterator(items[50:], compressionManager)

	reader1, err := sstableManger.CreateSSTable("test-table-1", iterator1, uint64(len(items)))
	if err != nil {
		panic("failed to create SSTable: " + err.Error())
	}
	reader2, err := sstableManger.CreateSSTable("test-table-2", iterator2, uint64(len(items)))

	if err != nil {
		panic("failed to create SSTable: " + err.Error())
	}

	merged, err := sstableManger.MergeSSTables("merged", []*sstable.SSTableReader{reader1, reader2})

	if err != nil {
		panic("failed to merge SSTables: " + err.Error())
	}

	n1, err := reader1.GetNumberOfItems()
	if err != nil {
		panic("failed to get number of items from reader1: " + err.Error())
	}
	n2, err := reader2.GetNumberOfItems()
	if err != nil {
		panic("failed to get number of items from reader2: " + err.Error())
	}
	m, err := merged.GetNumberOfItems()
	if err != nil {
		panic("failed to get number of items from merged reader: " + err.Error())
	}

	fmt.Printf("Reader1 has %d items, Reader2 has %d items, Merged has %d items.\n", n1, n2, m)

	for {
		var input string
		fmt.Print("Enter a key to search (or 'q' to quit): ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		if input == "q" {
			break
		}
		item, err := merged.Get(input)
		if err != nil {
			fmt.Println("Error getting item:", err)
			continue
		}
		if item == nil {
			fmt.Println("Key not found:", input)
		} else {
			fmt.Printf("Found item: Key=%s, Value=%s, Timestamp=%d, Deleted=%t\n",
				item.Key, string(item.Value), item.Timestamp, item.Tombstone)
		}
	}

	merge()
	// setupTestDir()
	// defer cleanupTestDir()

	// compressionManager := NewCompressionMockup(false)

	// dilutionFactor := uint8(10)

	// sstableManger := sstable.NewSSTableManager(testDir, sstable.MODE_MULTI_FILE, dilutionFactor, BM, compressionManager)

	// data := getData(150, compressionManager)

	// // Convert []*testItem to []sstable.SSTableItem
	// items := make([]sstable.ISSTableItem, len(data))
	// for i, v := range data {
	// 	items[i] = v
	// }

	// // Sort items by key
	// sort.Slice(items, func(i, j int) bool {
	// 	if !compressionManager.IsCompressionEnabled() {
	// 		return items[i].GetKey() < items[j].GetKey()
	// 	}
	// 	return items[i].GetCompressedKey() < items[j].GetCompressedKey()
	// })

	// // Create a test SSTable iterator
	// iterator := sstable.NewSSListIterator(items, compressionManager)

	// _, err := sstableManger.CreateSSTable("test-table", iterator, uint64(len(items)))
	// if err != nil {
	// 	panic("failed to create SSTable: " + err.Error())
	// }

	// reader := sstableManger.NewSSTableReader("test-table")

	// n, err := reader.GetNumberOfItems()
	// if err != nil {
	// 	panic("failed to get number of items: " + err.Error())
	// }

	// fmt.Printf("SSTable created with %d items.\n", n)
	// for {
	// 	var input string
	// 	fmt.Print("Enter a key to search (or 'q' to quit): ")
	// 	_, err := fmt.Scanln(&input)
	// 	if err != nil {
	// 		fmt.Println("Error reading input:", err)
	// 		continue
	// 	}
	// 	if input == "q" {
	// 		break
	// 	}
	// 	item, err := reader.Get(input)
	// 	if err != nil {
	// 		fmt.Println("Error getting item:", err)
	// 		continue
	// 	}
	// 	if item == nil {
	// 		fmt.Println("Key not found:", input)
	// 	} else {
	// 		fmt.Printf("Found item: Key=%s, Value=%s, Timestamp=%d, Deleted=%t\n",
	// 			item.Key, string(item.Value), item.Timestamp, item.Tombstone)
	// 	}
	// }

}
