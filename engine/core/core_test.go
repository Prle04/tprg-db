package core

import (
	"testing"
	externalconfiguration "tprgdb/engine/externalConfiguration"
)

func defaultConfig() *externalconfiguration.ExternalConfiguration {
	config := new(externalconfiguration.ExternalConfiguration)
	config.Validate()
	return config
}

func TestCoreSetAndGet(t *testing.T) {
	core, err := NewCore(defaultConfig())
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}

	key := "testkey"
	value := []byte("testvalue")

	if err := core.Set(key, value); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	item, err := core.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if item == nil || string(item.GetValue()) != "testvalue" || item.GetKey() != key {
		t.Errorf("Unexpected item: %+v", item)
	}
}

func TestCoreDelete(t *testing.T) {
	core, err := NewCore(defaultConfig())
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}

	key := "deletekey"
	value := []byte("deletevalue")

	if err := core.Set(key, value); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := core.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	item, err := core.Get(key)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if item != nil && !item.IsDeleted() {
		t.Errorf("Item should be deleted: %+v", item)
	}
}

func TestCoreGetMany(t *testing.T) {
	core, err := NewCore(defaultConfig())
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}

	keys := []string{"key1", "key2", "key3"}
	values := [][]byte{[]byte("val1"), []byte("val2"), []byte("val3")}

	for i, k := range keys {
		if err := core.Set(k, values[i]); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	items, err := core.GetMany(keys)
	if err != nil {
		t.Fatalf("GetMany failed: %v", err)
	}
	for i, item := range items {
		if item == nil || string(item.GetValue()) != string(values[i]) {
			t.Errorf("Unexpected item at %d: %+v", i, item)
		}
	}
}

func TestCoreGetKeysRange(t *testing.T) {
	core, err := NewCore(defaultConfig())
	if err != nil {
		t.Fatalf("Failed to create core: %v", err)
	}

	keys := []string{"a", "b", "c", "d", "e"}
	for _, k := range keys {
		if err := core.Set(k, []byte("v"+k)); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	result, err := core.GetKeysRange("b", "d", 0, 3)
	if err != nil {
		t.Fatalf("GetKeysRange failed: %v", err)
	}
	if len(result) == 0 {
		t.Errorf("Expected keys in range, got none")
	}
}
