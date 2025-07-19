package sstable_test

import (
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
	testDir := "test-temp-" + randomString(5)
	err := os.MkdirAll(testDir, 0o755)
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
		items[i].GetKey()           // Initialize key
		items[i].GetValue()         // Initialize value
		items[i].GetTimestamp()     // Initialize timestamp
		items[i].GetCompressedKey() // Initialize compressed key
		items[i].IsDeleted()        // Initialize tombstone
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
	if compressed, ok := c.s2i[s]; ok {
		return compressed, nil
	}
	c.count++
	c.s2i[s] = c.count
	c.i2s[c.count] = s
	return c.count, nil
}

func (c *compressionMockup) GetDecompressed(u uint64) (string, error) {
	if decompressed, ok := c.i2s[u]; ok {
		return decompressed, nil
	}
	panic(fmt.Sprintf("compressed value %d not found", u))
	// return "", fmt.Errorf("compressed value %d not found", u)
}

func (c *compressionMockup) IsCompressionEnabled() bool {
	return c.compressed
}

func generateItems(n int, compressionManager sstable.ICompressionManager) ([]sstable.ISSTableItem, []string) {
	data := getData(n, compressionManager)

	sort.Slice(data, func(i, j int) bool {
		return data[i].GetKey() < data[j].GetKey()
	})

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.ISSTableItem, len(data))
	keys := make([]string, len(data))
	for i, v := range data {
		items[i] = v
		keys[i] = v.GetKey()
	}
	return items, keys
}

func testCreate(t *testing.T, compression bool, mode string) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	compressionManager := NewCompressionMockup(compression)

	sstableManger := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	data := getData(150, compressionManager)

	sort.Slice(data, func(i, j int) bool {
		return data[i].GetKey() < data[j].GetKey()
	})

	// Convert []*testItem to []sstable.SSTableItem
	items := make([]sstable.ISSTableItem, len(data))
	keys := make([]string, len(data))
	for i, v := range data {
		items[i] = v
		keys[i] = v.GetKey()
	}

	// Create a test SSTable iterator
	iterator := sstable.NewSSListIterator(items, compressionManager)

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

func createTable(t *testing.T, compressionManager *compressionMockup, mode string, n int, testDir, tableName string) (*sstable.SSTableReader, []sstable.ISSTableItem, []string) {

	sstableManger := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	items, keys := generateItems(n, compressionManager)

	// Create a test SSTable iterator
	iterator := sstable.NewSSListIterator(items, compressionManager)

	reader, err := sstableManger.CreateSSTable(tableName, iterator, uint64(len(items)))
	if err != nil {
		t.Fatalf("failed to create SSTable: %v", err)
	}
	return reader, items, keys
}

func testIterator(t *testing.T, mode string, createCompressed, iterateCompressed bool) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	numberOfItems := 100
	compressionManager := NewCompressionMockup(createCompressed)

	_, _, keys := createTable(t, compressionManager, mode, numberOfItems, testDir, "test-table-iterator")

	compressionManager.compressed = iterateCompressed

	sstableManger := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	reader := sstableManger.NewSSTableReader("test-table-iterator")

	if reader == nil {
		t.Fatal("failed to create SSTable reader")
	}

	it := reader.GetIterator()
	if it == nil {
		t.Fatal("failed to create SSTable iterator")
	}

	count := 0
	for item, err := it.Next(); ; item, err = it.Next() {
		if err == blockmanager.EOF {
			break
		}
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		count++
		var key string
		if reader.IsCompressed() {
			key, _ = compressionManager.GetDecompressed(item.GetCompressedKey())
		} else {
			key = item.GetKey()
		}
		if key != keys[count-1] {
			t.Fatalf("expected key %s, got %s", keys[count-1], key)
		}
	}
	if count != numberOfItems {
		t.Fatalf("expected %d items, got %d", numberOfItems, count)
	}
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

func createMergeTest(t *testing.T, compressMerged bool, mergedMode string) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	compressionManager := NewCompressionMockup(compressMerged)

	sstableManger := sstable.NewSSTableManager(testDir, mergedMode, 10, BM, compressionManager)

	items, keys := generateItems(100, compressionManager)

	slices := splitWithOverlap(items, 5, 40)
	keyMap := make(map[string]int)
	for _, key := range keys {
		keyMap[key] = 0
	}
	readers := make([]*sstable.SSTableReader, len(slices))
	for i, slice := range slices {
		sort.Slice(slice, func(i, j int) bool {
			return slice[i].GetKey() < slice[j].GetKey()
		})
		tableName := fmt.Sprintf("test-table-merge-%d", i)
		reader, err := sstableManger.CreateSSTable(tableName, sstable.NewSSListIterator(slice, compressionManager), uint64(len(slice)))
		if err != nil {
			t.Fatalf("failed to create SSTable %s: %v", tableName, err)
		}
		if reader == nil {
			t.Fatalf("failed to create SSTable reader for %s", tableName)
		}
		readers[i] = reader
	}
	iter, err := sstable.NewSSTableMergeIterator(readers, compressMerged)
	if err != nil {
		t.Fatalf("failed to create SSTable merge iterator: %v", err)
	}
	for item, err := iter.Next(); ; item, err = iter.Next() {
		if err == blockmanager.EOF {
			break
		}
		if err != nil {
			t.Fatalf("merge iterator error: %v", err)
		}
		key := item.GetKey()
		value, exists := keyMap[key]
		if !exists {
			t.Fatalf("unexpected key found in merged SSTable: %s", key)
		}
		if value != 0 {
			t.Fatalf("key %s found multiple times in merged SSTable", key)
		}
		keyMap[key]++
	}
}

func testSSTableReaderValidate(t *testing.T, mode string) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	compressionManager := NewCompressionMockup(true)

	sstableManger := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	items, _ := generateItems(100, compressionManager)

	// Create a test SSTable iterator
	iterator := sstable.NewSSListIterator(items, compressionManager)

	_, err := sstableManger.CreateSSTable("test-table-validate", iterator, uint64(len(items)))
	if err != nil {
		t.Fatalf("failed to create SSTable: %v", err)
	}

	reader := sstableManger.NewSSTableReader("test-table-validate")
	if reader == nil {
		t.Fatal("failed to create SSTable reader")
	}

	valid, err := reader.Validate()
	if err != nil {
		t.Fatalf("failed to validate SSTable: %v", err)
	}
	if !valid {
		t.Fatal("SSTable validation failed")
	}
	file := BM.OpenBlockedFile(reader.GetFilePath() + "-data.db")
	file.Seek(sstable.SSTableHeaderSize)
	newData := make([]byte, 100)
	for i := range newData {
		newData[i] = byte(rand.Intn(256)) // Fill with some data
	}
	err = file.Write(newData)
	if err != nil {
		t.Fatalf("failed to write to SSTable file: %v", err)
	}
	err = file.Close()
	if err != nil {
		t.Fatalf("failed to close SSTable file: %v", err)
	}
	valid, err = reader.Validate()
	if err != nil {
		t.Fatalf("failed to validate SSTable after modification: %v", err)
	}
	if valid {
		t.Fatal("SSTable validation should have failed after modification")
	}
	reader.Delete() // Clean up the SSTable after test
}

func testGetMany(t *testing.T, compression bool, mode string) {
	itemsCount := 200
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	compressionManager := NewCompressionMockup(compression)
	sstableManager := sstable.NewSSTableManager(testDir, mode, 10, BM, compressionManager)

	items, keys := generateItems(itemsCount, compressionManager)
	iterator := sstable.NewSSListIterator(items, compressionManager)

	_, err := sstableManager.CreateSSTable("test-table-getmany", iterator, uint64(len(items)))
	if err != nil {
		t.Fatalf("failed to create SSTable: %v", err)
	}

	reader := sstableManager.NewSSTableReader("test-table-getmany")
	if reader == nil {
		t.Fatal("failed to create SSTable reader")
	}

	// Pick 50 random keys from the table and 50 random keys not in the table
	existingKeys := make([]string, 50)
	copy(existingKeys, keys[:50])
	nonExistingKeys := make([]string, 50)
	for i := 0; i < 50; i++ {
		nonExistingKeys[i] = randomString(10)
	}

	allKeys := append(existingKeys, nonExistingKeys...)

	sort.Slice(allKeys, func(i, j int) bool {
		return allKeys[i] < allKeys[j]
	})

	results, err := reader.GetMany(allKeys)
	if err != nil {
		t.Fatalf("failed to get many keys: %v", err)
	}

	if len(results) != len(allKeys) {
		t.Fatalf("expected %d results, got %d", len(allKeys), len(results))
	}

	for i, key := range allKeys {
		item := results[i]
		isExisting := slices.Contains(keys, key)
		if isExisting && item == nil {
			t.Fatalf("expected key %s to be found, but got nil", key)
		}
		if !isExisting && item != nil {
			t.Fatalf("expected key %s to not be found, but got a result", key)
		}
	}
}

// CREATE TABLE TESTS

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

// GET MANY TESTS
func TestGetManyWithCompressionOneFile(t *testing.T) {
	testGetMany(t, true, sstable.MODE_ONE_FILE)
}

func TestGetManyWithCompressionMultiFile(t *testing.T) {
	testGetMany(t, true, sstable.MODE_MULTI_FILE)
}

func TestGetManyWithoutCompressionOneFile(t *testing.T) {
	testGetMany(t, false, sstable.MODE_ONE_FILE)
}

func TestGetManyWithoutCompressionMultiFile(t *testing.T) {
	testGetMany(t, false, sstable.MODE_MULTI_FILE)
}

// ITERATOR TESTS

func TestIterateWithout2WithoutOneFile(t *testing.T) {
	testIterator(t, sstable.MODE_ONE_FILE, false, false)
}
func TestIterateWithout2WithOneFile(t *testing.T) {
	testIterator(t, sstable.MODE_ONE_FILE, false, true)
}

func TestIterateWith2WithoutOneFile(t *testing.T) {
	testIterator(t, sstable.MODE_ONE_FILE, true, false)
}

func TestIterateWith2WithOneFile(t *testing.T) {
	testIterator(t, sstable.MODE_ONE_FILE, true, true)
}

func TestIterateWithout2WithoutMultiFile(t *testing.T) {
	testIterator(t, sstable.MODE_MULTI_FILE, false, false)
}

func TestIterateWithout2WithMultiFile(t *testing.T) {
	testIterator(t, sstable.MODE_MULTI_FILE, false, true)
}
func TestIterateWith2WithoutMultiFile(t *testing.T) {
	testIterator(t, sstable.MODE_MULTI_FILE, true, false)
}
func TestIterateWith2WithMultiFile(t *testing.T) {
	testIterator(t, sstable.MODE_MULTI_FILE, true, true)
}

// MERGE TESTS
func TestMergeSSTablesWithCompressionOneFile(t *testing.T) {
	createMergeTest(t, true, sstable.MODE_ONE_FILE)
}
func TestMergeSSTablesWithCompressionMultiFile(t *testing.T) {
	createMergeTest(t, true, sstable.MODE_MULTI_FILE)
}
func TestMergeSSTablesWithoutCompressionOneFile(t *testing.T) {
	createMergeTest(t, false, sstable.MODE_ONE_FILE)
}
func TestMergeSSTablesWithoutCompressionMultiFile(t *testing.T) {
	createMergeTest(t, false, sstable.MODE_MULTI_FILE)
}

// VALIDATE TESTS
func TestSSTableReaderValidateOneFile(t *testing.T) {
	testSSTableReaderValidate(t, sstable.MODE_ONE_FILE)
}
func TestSSTableReaderValidateMultiFile(t *testing.T) {
	testSSTableReaderValidate(t, sstable.MODE_MULTI_FILE)
}
