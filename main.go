package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

var testDir string

var BM = blockmanager.NewBlockManager(2048, 128)

func setupTestDir() {
	id := rand.Int63()
	testDir = filepath.Join("test-temp-" + strconv.FormatInt(id, 10))
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
	setupTestDir()
	// defer cleanupTestDir()

	data := getData(100)

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.SSTableItem, len(data))
	for i, v := range data {
		items[i] = v
	}

	err := sstable.FlushToSSTable(items, testDir, "test-table", 10, "multi-file", BM)
	if err != nil {
		panic("failed to create SSTable: " + err.Error())
	}
}
