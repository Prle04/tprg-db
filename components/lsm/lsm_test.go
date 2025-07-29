package lsm

//import (
//	"fmt"
//	"strconv"
//	"testing"
//	"tprgdb/components/blockmanager"
//	"tprgdb/components/sstable"
//)
//
//type mockItem struct {
//	key           string
//	value         []byte
//	compressedKey uint64
//	timestamp     uint64
//	tombstone     bool
//}
//
//func (m *mockItem) GetKey() string           { return m.key }
//func (m *mockItem) GetValue() []byte         { return m.value }
//func (m *mockItem) GetCompressedKey() uint64 { return m.compressedKey }
//func (m *mockItem) GetTimestamp() uint64     { return m.timestamp }
//func (m *mockItem) IsDeleted() bool          { return m.tombstone }
//
//type compressionMockup struct {
//	s2i        map[string]uint64
//	i2s        map[uint64]string
//	count      uint64
//	compressed bool
//}
//
//func NewCompressionMockup(compressed bool) *compressionMockup {
//	return &compressionMockup{
//		s2i:        make(map[string]uint64),
//		i2s:        make(map[uint64]string),
//		count:      0,
//		compressed: compressed,
//	}
//}
//
//func (c *compressionMockup) GetCompressed(s string) (uint64, error) {
//	if compressed, ok := c.s2i[s]; ok {
//		return compressed, nil
//	}
//	c.count++
//	c.s2i[s] = c.count
//	c.i2s[c.count] = s
//	return c.count, nil
//}
//
//func (c *compressionMockup) GetDecompressed(u uint64) (string, error) {
//	if decompressed, ok := c.i2s[u]; ok {
//		return decompressed, nil
//	}
//	return "", fmt.Errorf("compressed value %d not found", u)
//}
//
//func (c *compressionMockup) IsCompressionEnabled() bool {
//	return c.compressed
//}
//
//func TestLSMAddAndGet(t *testing.T) {
//	tmpDir := t.TempDir()
//
//	blockMgr := blockmanager.NewBlockManager(2048, 128)
//	compression := NewCompressionMockup(true) // Use mock compression manager
//
//	sstMgr := sstable.NewSSTableManager(tmpDir, "one-file", 2, blockMgr, compression)
//
//	tree := NewLSM(3, tmpDir, SIZE_TIERED, blockMgr, sstMgr, compression, 2, 10)
//	if tree == nil {
//		t.Fatal("Failed to create LSM tree")
//	}
//
//	// Add 10 items
//	for i := 0; i < 10; i++ {
//		key := "key" + strconv.Itoa(i)
//		value := []byte("value" + strconv.Itoa(i))
//
//		item := &mockItem{
//			key:           key,
//			value:         value,
//			compressedKey: uint64(1000 + i),
//			timestamp:     uint64(i + 1),
//			tombstone:     false,
//		}
//
//		err := tree.AddTable([]sstable.ISSTableItem{item})
//		if err != nil {
//			t.Fatalf("AddTable failed: %v", err)
//		}
//	}
//
//	// Verify retrieval
//	for i := 0; i < 10; i++ {
//		key := "key" + strconv.Itoa(i)
//		got, err := tree.Get(key)
//		if err != nil {
//			t.Errorf("Get(%s) error: %v", key, err)
//			continue
//		}
//		if got == nil {
//			t.Errorf("Expected value for key %s but got nil", key)
//			continue
//		}
//		expectedVal := "value" + strconv.Itoa(i)
//		if string(got.GetValue()) != expectedVal {
//			t.Errorf("For key %s expected value %s but got %s", key, expectedVal, string(got.GetValue()))
//		}
//	}
//}

import (
	"sort"
	"strconv"
	"testing"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

// --- mockItem as you have it ---
type mockItem struct {
	key           string
	value         []byte
	compressedKey uint64
	timestamp     uint64
	tombstone     bool
}

func (m *mockItem) GetKey() string           { return m.key }
func (m *mockItem) GetValue() []byte         { return m.value }
func (m *mockItem) GetCompressedKey() uint64 { return m.compressedKey }
func (m *mockItem) GetTimestamp() uint64     { return m.timestamp }
func (m *mockItem) IsDeleted() bool          { return m.tombstone }

// --- compressionMockup as you have it ---
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
	return "", nil
}

func (c *compressionMockup) IsCompressionEnabled() bool {
	return c.compressed
}

// --- Helper to create a new test LSM tree ---
func newTestLSM(t *testing.T, maxLevel int, algo string, filePerLevel int, multiplier int) *LSM {
	tmpDir := t.TempDir()
	blockMgr := blockmanager.NewBlockManager(2048, 128)
	compression := NewCompressionMockup(true)
	sstMgr := sstable.NewSSTableManager(tmpDir, "one-file", 2, blockMgr, compression)

	tree := NewLSM(maxLevel, tmpDir, algo, blockMgr, sstMgr, compression, filePerLevel, multiplier)
	if tree == nil {
		t.Fatal("Failed to create LSM tree")
	}
	return tree
}

// --- Test AddTable and Get with simple items ---
func TestLSMAddAndGet(t *testing.T) {
	tree := newTestLSM(t, 3, SIZE_TIERED, 2, 10)

	for i := 0; i < 10; i++ {
		item := &mockItem{
			key:           "key" + strconv.Itoa(i),
			value:         []byte("value" + strconv.Itoa(i)),
			compressedKey: uint64(1000 + i),
			timestamp:     uint64(i + 1),
		}

		err := tree.AddTable([]sstable.ISSTableItem{item})
		if err != nil {
			t.Fatalf("AddTable failed: %v", err)
		}
	}

	for i := 0; i < 10; i++ {
		got, err := tree.Get("key" + strconv.Itoa(i))
		if err != nil {
			t.Errorf("Get error for key%d: %v", i, err)
			continue
		}
		if got == nil {
			t.Errorf("Expected value for key%d but got nil", i)
			continue
		}
		expected := "value" + strconv.Itoa(i)
		if string(got.GetValue()) != expected {
			t.Errorf("Expected %s, got %s", expected, string(got.GetValue()))
		}
	}
}

// --- Test that compaction triggers in size-tiered mode ---
func TestSizeTierCompaction(t *testing.T) {
	tree := newTestLSM(t, 3, SIZE_TIERED, 2, 10)

	// Add 3 tables to level 0 which should trigger compaction (filePerLevel=2)
	for i := 0; i < 3; i++ {
		item := &mockItem{
			key:           "k" + strconv.Itoa(i),
			value:         []byte("v" + strconv.Itoa(i)),
			compressedKey: uint64(i),
			timestamp:     uint64(i),
		}
		err := tree.AddTable([]sstable.ISSTableItem{item})
		if err != nil {
			t.Fatalf("AddTable failed: %v", err)
		}
	}

	// After compaction, level 0 should have <= 2 tables, level 1 should have merged tables
	if len(tree.tables[0]) > 2 {
		t.Errorf("Expected level 0 to have <= 2 tables, got %d", len(tree.tables[0]))
	}
	if len(tree.tables[1]) == 0 {
		t.Errorf("Expected level 1 to have merged tables")
	}
}

// --- Test that compaction triggers in leveled mode ---
func TestLeveledCompaction(t *testing.T) {
	tree := newTestLSM(t, 3, LEVELED, 1, 2)

	// Add 3 tables to level 0; filePerLevel=1, multiplier=2 so compaction should trigger
	for i := 0; i < 3; i++ {
		item := &mockItem{
			key:           "k" + strconv.Itoa(i),
			value:         []byte("v" + strconv.Itoa(i)),
			compressedKey: uint64(i),
			timestamp:     uint64(i),
		}
		err := tree.AddTable([]sstable.ISSTableItem{item})
		if err != nil {
			t.Fatalf("AddTable failed: %v", err)
		}
	}

	// Check that compaction occurred: level 0 tables <= limit, level 1 has tables
	if len(tree.tables[0]) > 1 {
		t.Errorf("Expected level 0 to have <= 1 table, got %d", len(tree.tables[0]))
	}
	if len(tree.tables[1]) == 0 {
		t.Errorf("Expected level 1 to have merged tables")
	}
}

// --- Test Save and Load ---
func TestSaveLoad(t *testing.T) {
	tree := newTestLSM(t, 3, SIZE_TIERED, 2, 10)

	item := &mockItem{
		key:           "foo",
		value:         []byte("bar"),
		compressedKey: 42,
		timestamp:     1,
	}

	if err := tree.AddTable([]sstable.ISSTableItem{item}); err != nil {
		t.Fatalf("AddTable failed: %v", err)
	}

	err := tree.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new tree, load from disk
	newTree := LSM{
		sstableManager:      tree.sstableManager,
		compressionManager:  tree.compressionManager,
		blockManager:        tree.blockManager,
		folderPath:          tree.folderPath,
		tables:              make([][]*sstable.SSTableReader, tree.maxLevel),
		maxLevel:            tree.maxLevel,
		compactionAlgorithm: tree.compactionAlgorithm,
		filePerLevel:        tree.filePerLevel,
		multiplier:          tree.multiplier,
	}
	newTree.folderPath = tree.folderPath // Use same folder path for loading
	err = newTree.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	got, err := newTree.Get("foo")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatalf("Expected value for key 'foo', got nil")
	}
	if string(got.GetValue()) != "bar" {
		t.Fatalf("Expected 'bar', got %s", string(got.GetValue()))
	}
}

// --- Test getting a key that does not exist returns nil without error ---
func TestGetMissingKey(t *testing.T) {
	tree := newTestLSM(t, 3, SIZE_TIERED, 2, 10)
	got, err := tree.Get("missing")
	if err != nil {
		t.Errorf("Expected no error for missing key, got %v", err)
	}
	if got != nil {
		t.Errorf("Expected nil for missing key, got %+v", got)
	}
}

// --- Test GetMany with multiple keys ---
func TestGetMany(t *testing.T) {
	tree := newTestLSM(t, 3, SIZE_TIERED, 2, 10)

	// Add some items
	for i := 0; i < 5; i++ {
		item := &mockItem{
			key:           "key" + strconv.Itoa(i),
			value:         []byte("value" + strconv.Itoa(i)),
			compressedKey: uint64(1000 + i),
			timestamp:     uint64(i + 1),
		}
		err := tree.AddTable([]sstable.ISSTableItem{item})
		if err != nil {
			t.Fatalf("AddTable failed: %v", err)
		}
	}

	keys := []string{"key1", "key3", "missing"}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	results, err := tree.GetMany(keys)
	if err != nil {
		t.Fatalf("GetMany failed: %v", err)
	}

	if len(results) != len(keys) {
		t.Fatalf("Expected %d results, got %d", len(keys), len(results))
	}

	expectedValues := map[string]string{
		"key1": "value1",
		"key3": "value3",
	}

	for i, key := range keys {
		if result := results[i]; result != nil {
			if expected, ok := expectedValues[key]; ok {
				if string(result.GetValue()) != expected {
					t.Errorf("For key %s expected value %s but got %s", key, expected, string(result.GetValue()))
				}
			} else {
				t.Errorf("Unexpected result for key %s: %+v", key, result)
			}
		} else if _, ok := expectedValues[key]; ok {
			t.Errorf("Expected value for key %s but got nil", key)
		}
	}
}
