package sstable_test

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"sort"
	"testing"
	"time"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

var BM = blockmanager.NewBlockManager(2048, 128)

func setupTestDir(t *testing.T) string {
	testDir := "test-temp" + randomString(5)
	err := os.Mkdir(testDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	return testDir
}

func cleanupTestDir(t *testing.T, testDir string) {
	if testDir != "" {
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Fatalf("failed to remove test dir: %v", err)
		}
		testDir = ""
	}
}

type testItem struct {
	Key           *string
	CompressedKey *uint64
	Value         []byte
	Timestamp     *uint64
	Tombstone     *bool
	CM            sstable.ICompressionManager
}

func (item *testItem) GetCompressedKey() uint64 {
	if item == nil {
		panic("ITEM IS NILLLLLLL")
	}
	if item.CompressedKey != nil {
		return *item.CompressedKey
	}
	item.CompressedKey = new(uint64)
	*item.CompressedKey, _ = item.CM.GetCompressed(item.GetKey())
	return *item.CompressedKey
}

func (item *testItem) GetKey() string {
	if item == nil {
		panic("ITEM IS NILLLLLLL")
	}
	if item.Key != nil {
		return *item.Key
	}
	item.Key = new(string)
	*item.Key = randomString(10)
	return *item.Key
}

func (item *testItem) GetValue() []byte {
	if item.Value != nil {
		return item.Value
	}
	length := rand.Intn(100) + 1 // Random length between 1 and 100
	str := randomString(length)
	item.Value = []byte(str)
	return []byte(str)
}

func (item *testItem) GetTimestamp() uint64 {
	if item.Timestamp != nil {
		return *item.Timestamp
	}
	item.Timestamp = new(uint64)
	*item.Timestamp = uint64(time.Now().Unix())
	return *item.Timestamp
}

func (item *testItem) IsDeleted() bool {
	if item.Tombstone != nil {
		return *item.Tombstone
	}
	item.Tombstone = new(bool)
	*item.Tombstone = rand.Intn(2) == 0 // Randomly return true or false
	return *item.Tombstone
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getData(n int, compressionManager sstable.ICompressionManager) []*testItem {
	items := make([]*testItem, n)
	for i := range n {
		items[i] = &testItem{
			Key:       nil,
			Value:     nil,
			Timestamp: nil,
			Tombstone: nil,
			CM:        compressionManager,
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

func testCreate(t *testing.T, compression bool, mode string) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	compressionManager := NewCompressionMockup(compression)

	sstableManger := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	data := getData(150, compressionManager)

	sort.Slice(data, func(i, j int) bool {
		if !compressionManager.IsCompressionEnabled() {
			return data[i].GetKey() < data[j].GetKey()
		}
		return data[i].GetCompressedKey() < data[j].GetCompressedKey()
	})

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.ISSTableItem, len(data))
	keys := make([]string, len(data))
	for i, v := range data {
		items[i] = v
		keys[i] = v.GetKey()
	}

	// Create a test SSTable iterator
	iterator := &testSSTableIterator{
		items: items,
		index: 0,
	}

	_, err := sstableManger.CreateSSTable("test-table", iterator, uint64(len(items)))
	if err != nil {
		t.Fatalf("failed to create SSTable: %v", err)
	}

	reader := sstableManger.NewSSTableReader("test-table")
	if reader == nil {
		t.Fatal("failed to create SSTable reader")
	}

	n, err := reader.GetNumberOfItems()
	if err != nil {
		t.Fatalf("failed to get number of items: %v", err)
	}
	if n != uint64(len(data)) {
		t.Fatalf("expected %d items, got %d", len(data), n)
	}
	for _, item := range items {
		read, err := reader.Get(item.GetKey())
		if err != nil {
			t.Fatalf("failed to get item %s: %v", item.GetKey(), err)
		}
		if read == nil {
			t.Fatalf("item %s not found in SSTable", item.GetKey())
		}
		if compressionManager.IsCompressionEnabled() {
			if read.GetCompressedKey() != item.GetCompressedKey() {
				t.Fatalf("expected compressed key %d, got %d", item.GetCompressedKey(), read.GetCompressedKey())
			}
		} else {
			if read.GetKey() != item.GetKey() {
				t.Fatalf("expected key %s, got %s", item.GetKey(), read.GetKey())
			}
		}
		// if string(read.GetValue()) != string(item.GetValue()) {
		// 	t.Fatalf("expected value %s, got %s", string(item.GetValue()), string(read.GetValue()))
		// }
		if read.GetTimestamp() != item.GetTimestamp() {
			t.Fatalf("expected timestamp %d, got %d", item.GetTimestamp(), read.GetTimestamp())
		}
		if read.IsDeleted() != item.IsDeleted() {
			t.Fatalf("expected tombstone %t, got %t", item.IsDeleted(), read.IsDeleted())
		}
	}
	for i := 0; i < 100; i++ {
		key := randomString(10)
		read, err := reader.Get(key)
		isIn := slices.Contains(keys, key)
		if err != nil {
			t.Fatalf("failed to get item %s: %v", key, err)
		}
		if isIn && read == nil {
			t.Fatalf("item %s should be found in SSTable but was not", key)
		}
		if !isIn && read != nil {
			t.Fatalf("item %s should not be found in SSTable but was", key)
		}
	}
}

func TestCreateSSTableWithCompressionOneFile(t *testing.T) {
	testCreate(t, true, sstable.MODE_ONE_FILE)
}
func TestCreateSSTableWithCompressionMultipleFiles(t *testing.T) {
	testCreate(t, true, sstable.MODE_MULTI_FILE)
}
func TestCreateSSTableWithoutCompression(t *testing.T) {
	testCreate(t, false, sstable.MODE_ONE_FILE)
}
func TestCreateSSTableWithoutCompressionMultipleFiles(t *testing.T) {
	testCreate(t, false, sstable.MODE_MULTI_FILE)
}
