package bloomfilter_test

import (
	"testing"
	"tprgdb/components/bloomfilter"
)

func TestBloomFilter(t *testing.T) {
	bf := bloomfilter.NewBloomFilter(100, 0.01)
	data := []byte("data")
	bf.Insert(data)
	if bf.IsNotIn(data) {
		t.Errorf("Expected data to be in bloom filter")
	}
	if !bf.IsNotIn([]byte("not in")) {
		t.Errorf("Expected data not to be in bloom filter")
	}
}

func TestSerialization(t *testing.T) {
	bf := bloomfilter.NewBloomFilter(100, 0.01)
	data := []byte("data")
	bf.Insert(data)
	bytes := bf.SerializeToBytes()
	var bf2 bloomfilter.BloomFilter
	bf2.DeserializeFromBytes(bytes)
	if bf2.IsNotIn(data) {
		t.Errorf("Expected data to be in bloom filter")
	}
	if !bf2.IsNotIn([]byte("not in")) {
		t.Errorf("Expected data not to be in bloom filter")
	}
}
