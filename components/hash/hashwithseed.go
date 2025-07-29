package hash

import (
	"crypto/md5"
	"time"
	"tprgdb/components/serialization"
)

type HashWithSeed struct {
	Seed []byte // 32 bitni broj - 4byte
}

func (h HashWithSeed) Hash(data []byte) uint64 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return serialization.BytesToUInt64(fn.Sum(nil))
}

func CreateHashFunctions(k uint32) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint32(time.Now().Unix())
	for i := uint32(0); i < k; i++ {
		seed := serialization.UInt32ToBytes(ts + i)
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}

func GetMD5Hash(data []byte) uint64 {
	h := md5.Sum(data)
	return serialization.BytesToUInt64(h[:])
}
