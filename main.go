package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

var testDir string = "test-temp"

var BM = blockmanager.NewBlockManager(2048, 128)

func setupTestDir() {
	testDir = filepath.Join("test-temp")
	err := os.Mkdir(testDir, 0o755)
	if err != nil {
		panic(err.Error())
	}
}

func cleanupTestDir() {
	if testDir != "" {
		err := os.RemoveAll(testDir)
		if err != nil {
			panic("failed to remove test dir: " + err.Error())
		}
		testDir = ""
	}
}

type testItem struct {
	key   string
	value []byte
}

func (item testItem) GetKey() string {
	return item.key
	// return randomString(10)
}

func (item testItem) GetValue() []byte {
	return item.value
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

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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

func main() {
	// setupTestDir()
	// defer cleanupTestDir()

	data := getData(100)

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.SSTableItem, len(data))
	for i, v := range data {
		items[i] = v
	}

	// Sort items by key
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetKey() < items[j].GetKey()
	})

	err := sstable.FlushToSSTable(items, testDir, "test-table", 10, "one-file", BM)
	if err != nil {
		panic("failed to create SSTable: " + err.Error())
	}

	reader := sstable.NewSSTableReader(testDir, "test-table", BM)

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