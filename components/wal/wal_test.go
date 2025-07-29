package wal_test

import (
	"fmt"
	"testing"
	"tprgdb/components/blockmanager"
	"tprgdb/components/wal"
)

func TestWAL(t *testing.T) {
	blockSize := uint64(256)
	numOfBlocks := uint64(2)
	paddingThreshold := uint64(10)

	// Create a temporary file for testing
	tempDir := t.TempDir()

	BM := blockmanager.NewBlockManager(blockSize, numOfBlocks)

	WAL, err := wal.NewWriteAheadLog(tempDir, numOfBlocks, paddingThreshold, BM)
	if err != nil {
		t.Fatalf("Failed to create WriteAheadLog: %v", err)
	}

	for i := uint64(0); i < 100; i++ {
		key := fmt.Sprintf("key%4d", i)
		value := []byte(fmt.Sprintf("value%4d", i))
		if i%15 == 0 {
			if err := WAL.Delete(key); err != nil {
				t.Fatalf("Failed to delete key %s: %v", key, err)
			}
		} else {
			if err := WAL.Add(key, value); err != nil {
				t.Fatalf("Failed to add key %s: %v", key, err)
			}
		}
	}
	// Close the WAL to ensure all data is written
	if err := WAL.Close(); err != nil {
		t.Fatalf("Failed to close WriteAheadLog: %v", err)
	}
	// Reopen the WAL to verify data integrity
	// WAL, err = wal.NewWriteAheadLog(tempDir, numOfBlocks, paddingThreshold, BM)
	// if err != nil {
	// 	t.Fatalf("Failed to reopen WriteAheadLog: %v", err)
	// }
	i := 0
	for it := range WAL.Iterate() {
		// t.Fatalf("ITERATOR CALLED")
		if it == nil {
			t.Fatal("Iterator returned nil")
		}
		if it.Err != nil {
			t.Fatalf("Iterator error: %v", it.Err)
		}
		key := fmt.Sprintf("key%4d", i)
		value := []byte(fmt.Sprintf("value%4d", i))
		if it.Record.GetKey() != key {
			t.Fatalf("Iterator returned unexpected key: got %s, want %s", it.Record.GetKey(), key)
		}
		if i%15 == 0 {
			if it.Record.GetValue() != nil {
				t.Fatal("Iterator returned value for deleted key")
			}
			if !it.Record.GetTombstone() {
				t.Fatalf("Iterator returned non-deleted record for key %s", key)
			}
		} else {
			if string(it.Record.GetValue()) != string(value) {
				t.Fatalf("Iterator returned unexpected value for key %s: got %s, want %s", key, it.Record.GetValue(), value)
			}
		}
		i++
	}
	if i != 100 {
		t.Fatalf("Iterator did not return expected number of records: got %d, want 100", i)
	}
}
