package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

var testDir string = "test-temp"

var BM = blockmanager.NewBlockManager(2048, 128)

// func setupTestDir() {
// 	testDir = filepath.Join("test-temp")
// 	err := os.Mkdir(testDir, 0o755)
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
	key   string
	value []byte
}

func (item testItem) GetKey() string {
	return item.key
	// return randomString(10)
}

func (item testItem) GetValue() []byte {
	return []byte(strings.Repeat(item.key, 100))
	// length := rand.Intn(100) + 1 // Random length between 1 and 100
	// str := randomString(length)
	// return []byte(str)
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

func getData(n int) []*testItem {
	items := make([]*testItem, n)
	for i := 0; i < n; i++ {
		items[i] = &testItem{
			key:   fmt.Sprintf("key-%d", i),
			value: []byte(fmt.Sprintf("value-%d", i)),
		}
	}
	return items
}

type testSSTableIterator struct {
	items []sstable.ISSTableItem
	index int
}

func (iter *testSSTableIterator) Next() (sstable.ISSTableItem, error) {
	if iter.index >= len(iter.items) {
		return nil, errors.New("kraj") // No more items
	}
	item := iter.items[iter.index]
	iter.index++
	return item, nil
}

type compressionMockup struct {
	compressed bool
}

func (c *compressionMockup) GetCompressied(s string) ([]byte, error) {
	if !c.compressed {
		return []byte(s), nil
	}
	// Mock compression: just reverse the string
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return []byte(string(runes)), nil
}

func (c *compressionMockup) GetDecompressied(u uint64) (string, error) {
	// Mock decompression: just convert uint64 to string
	return fmt.Sprintf("%d", u), nil
}

func (c *compressionMockup) IsCompressionEnabled() bool {
	return c.compressed
}

func main() {
	// setupTestDir()
	// defer cleanupTestDir()

	compressionManager := &compressionMockup{compressed: false}

	dilutionFactor := uint8(10)

	sstableManger := sstable.NewSSTableManager(testDir, sstable.MODE_ONE_FILE, dilutionFactor, BM, compressionManager)

	data := getData(150)

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.ISSTableItem, len(data))
	for i, v := range data {
		items[i] = v
	}

	// Sort items by key
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetKey() < items[j].GetKey()
	})

	// Create a test SSTable iterator
	iterator := &testSSTableIterator{
		items: items,
		index: 0,
	}

	_, err := sstableManger.CreateSSTable("test-table", iterator, uint64(len(items)))
	if err != nil {
		panic("failed to create SSTable: " + err.Error())
	}

	reader := sstableManger.NewSSTableReader("test-table")

	n, err := reader.GetNumberOfItems()
	if err != nil {
		panic("failed to get number of items: " + err.Error())
	}

	fmt.Printf("SSTable created with %d items.\n", n)
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
		item, err := reader.Get(input)
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

}
