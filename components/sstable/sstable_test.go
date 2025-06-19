package sstable_test

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

var testDir string

var BM = blockmanager.NewBlockManager(2048, 128)

func setupTestDir(t *testing.T) {
	id := rand.Int63()
	testDir = filepath.Join(os.TempDir(), "test-temp-"+strconv.FormatInt(id, 10))
	err := os.Mkdir(testDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
}

func cleanupTestDir(t *testing.T) {
	if testDir != "" {
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Fatalf("failed to remove test dir: %v", err)
		}
		testDir = ""
	}
}

type testItem struct {
}

func (item testItem) GetKey() string {
	return randomString(10)
}

func (item testItem) GetValue() []byte {
	length := rand.Intn(100) + 1 // Random length between 1 and 100
	str := randomString(length)
	return []byte(str)
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
		items[i] = &testItem{}
	}
	return items
}

func TestSSTable_CreateSSTable(t *testing.T) {
	setupTestDir(t)
	defer cleanupTestDir(t)

	data := getData(1_000)

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.SSTableItem, len(data))
	for i, v := range data {
		items[i] = v
	}

	err := sstable.FlushToSSTable(items, testDir, "test-table", 10, "one-file", BM)
	if err != nil {
		t.Fatalf("failed to create SSTable: %v", err)
	}
}
