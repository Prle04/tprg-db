package main

import (
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
	panic(fmt.Sprintf("Compressed value %d not found", u))
	return "", fmt.Errorf("compressed value %d not found", u)
}

func (c *compressionMockup) IsCompressionEnabled() bool {
	return c.compressed
}

func main() {
	// setupTestDir()
	// defer cleanupTestDir()

	compressionManager := NewCompressionMockup(true)

	dilutionFactor := uint8(10)

	sstableManger := sstable.NewSSTableManager(testDir, sstable.MODE_MULTI_FILE, dilutionFactor, BM, compressionManager)

	data := getData(150, compressionManager)

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

	// Create a test SSTable iterator
	iterator := sstable.NewSSListIterator(items, compressionManager)

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
