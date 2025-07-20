package avltree_test

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
	"tprgdb/components/avltree"
	"tprgdb/components/blockmanager"
)

func TestAVLTree_InsertFind(t *testing.T) {
	tree := avltree.CreateAVLTree()
	tree.Insert("apple", 10)
	tree.Insert("banana", 20)
	tree.Insert("cherry", 30)

	if tree.GetNumberOfElements() != 3 {
		t.Errorf("Expected 3 elements, got %d", tree.GetNumberOfElements())
	}

	data := tree.Find("banana")
	if data == 0 || data != 20 {
		t.Errorf("Expected value 20 for key 'banana', got %v", data)
	}

	notFound := tree.Find("durian")
	if notFound != 0 {
		t.Errorf("Expected nil for non-existent key, got %v", notFound)
	}
}

func TestSerializeDeserializeBytes(t *testing.T) {
	original := avltree.CreateAVLTree()
	original.Insert("alpha", 1)
	original.Insert("beta", 2)
	original.Insert("gamma", 3)

	data := original.SerializeToBytes()

	restored := avltree.CreateAVLTree()
	restored.DeserializeFromBytes(data)

	if restored.GetNumberOfElements() != 3 {
		t.Errorf("Expected 3 elements, got %d", restored.GetNumberOfElements())
	}

	check := []struct {
		key   string
		value uint64
	}{
		{"alpha", 1},
		{"beta", 2},
		{"gamma", 3},
	}

	for _, item := range check {
		data := restored.Find(item.key)
		if data == 0 || data != item.value {
			t.Errorf("Expected key %s with value %d, got %v", item.key, item.value, data)
		}
	}
}

func TestSimpleSerializeDeserialize(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	original := avltree.CreateAVLTree()
	expected := make(map[string]uint64)

	for i := 0; i < 100; i++ {
		key := randomString(10)
		value := uint64(rand.Intn(1000) + 1)
		original.Insert(key, value)
		expected[key] = value
	}

	serialized := original.SerializeToBytes()
	restored := avltree.CreateAVLTree()
	restored.DeserializeFromBytes(serialized)

	for k, v := range expected {
		data := restored.Find(k)
		if data == 0 || data != v {
			t.Errorf("Expected key %s with value %d, got %v", k, v, data)
		}
	}
}

// randomString generiše nasumičan string dužine n
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestSerializeDeserializeToFile_Random(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// 1. Креирамо оригинално стабло и насумичне податке
	original := avltree.CreateAVLTree()
	expected := make(map[string]uint64)

	for i := 0; i < 100; i++ {
		key := randomString(10)
		value := uint64(rand.Intn(1000) + 1)
		original.Insert(key, value)
		expected[key] = value
	}

	// 2. Отварамо blocked file
	tmpFile := filepath.Join(os.TempDir(), "avltree_test_data.dat")
	manager := blockmanager.NewBlockManager(64, 4096)
	file := manager.OpenBlockedFile(tmpFile)
	defer file.Delete()

	// 3. Сачувамо у фајл
	err := original.SerializeToFile(file)
	if err != nil {
		t.Fatalf("Error during serialization to file: %v", err)
	}

	// 4. Ресетујемо позицију
	file.Seek(0)

	// 5. Учитавамо из фајла
	restored := avltree.CreateAVLTree()
	err = restored.DeserializeFromFile(file)
	if err != nil {
		t.Fatalf("Error during deserialization from file: %v", err)
	}

	// 6. Проверавамо све вредности
	for key, expectedValue := range expected {
		data := restored.Find(key)
		if data == 0 {
			t.Errorf("Missing key after deserialization: %s", key)
			continue
		}
		if data != expectedValue {
			t.Errorf("Wrong value for key %s: expected %d, got %d", key, expectedValue, data)
		}
	}
}
