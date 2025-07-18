package hyperloglog

import (
	"fmt"
	"io"
	"math"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type HyperLogLog struct {
	p, m    uint64
	buckets []uint8
	hashFun hash.HashWithSeed
}

const CURRENT_VERSION uint8 = 1

func New(bits uint64) HyperLogLog {
	hll := HyperLogLog{
		p:       bits,
		m:       1 << bits,
		buckets: make([]uint8, 1<<bits),
		hashFun: hash.CreateHashFunctions(1)[0],
	}
	return hll
}

func countTrailingZeros(v uint64) uint8 {

	if v == 0 {
		return 64
	}

	var (
		x     uint64 = 1
		count uint8  = 0
	)

	for {
		rez := v & x
		if rez != 0 {
			break
		}
		x = x << 1
		count++
	}
	return count
}

func (h *HyperLogLog) Insert(data []byte) {

	hashValue := h.hashFun.Hash(data)
	index := hashValue >> (64 - h.p)
	leadingZeros := countTrailingZeros(hashValue) + 1 // Pomeri da ne računa indeks deo
	if h.buckets[index] < leadingZeros {
		h.buckets[index] = leadingZeros
	}

}

func (h HyperLogLog) Estimate() float64 {
	m := float64(h.m)
	sum := 0.0
	for _, b := range h.buckets {
		sum += 1.0 / math.Pow(2, float64(b))
	}

	alpha := 0.7213 / (1 + 1.079/m)
	estimate := alpha * m * m / sum

	return estimate
}

func (h HyperLogLog) Save(file io.Writer) error {
	bytes := serialization.UInt8ToBytes(CURRENT_VERSION)
	n, err := file.Write(bytes)
	if err != nil {
		fmt.Println("failed to write version", err)
		return err
	}
	if n != len(bytes) {
		return fmt.Errorf("saved only %d, expected %d", n, len(bytes))
	}

	bytes = serialization.UInt64ToBytes(h.p)
	n, err = file.Write(bytes)
	if err != nil {
		fmt.Println("failed to write p", err)
		return err
	}
	if n != len(bytes) {
		return fmt.Errorf("saved only %d, expected %d", n, len(bytes))
	}

	n, err = file.Write(h.hashFun.Seed)
	if err != nil {
		fmt.Println("failed to write seed", err)
		return err
	}
	if n != len(h.hashFun.Seed) {
		return fmt.Errorf("saved only %d, expected %d", n, len(h.hashFun.Seed))
	}

	for _, reg := range h.buckets {
		n, err = file.Write(serialization.UInt8ToBytes(reg))
		if err != nil {
			fmt.Println("failed to write bucket", err)
			return err
		}
		if n != 1 {
			return fmt.Errorf("saved only %d, expected 1", n)
		}
	}

	return nil
}

func (h *HyperLogLog) Load(file io.Reader) error {
	buffer1 := [1]byte{}
	n, err := file.Read(buffer1[:])
	if err != nil {
		fmt.Println("failed to read version", err)
		return err
	}
	if n != len(buffer1) {
		return fmt.Errorf("read only %d, expected %d", n, len(buffer1))
	}
	version := serialization.BytesToUInt8(buffer1[:])
	if version != CURRENT_VERSION {
		return fmt.Errorf("got version %d, expected %d", version, CURRENT_VERSION)
	}

	buffer8 := [8]byte{}
	n, err = file.Read(buffer8[:])
	if err != nil {
		fmt.Println("failed to read p", err)
		return err
	}
	if n != len(buffer8) {
		return fmt.Errorf("read only %d, expected %d", n, len(buffer8))
	}
	h.p = serialization.BytesToUInt64(buffer8[:])

	h.m = 1 << h.p

	seed := make([]byte, 4)
	n, err = file.Read(seed)
	if err != nil {
		fmt.Println("failed to read hash seed", err)
		return err
	}
	if n != len(seed) {
		return fmt.Errorf("read only %d, expected %d", n, len(seed))
	}
	h.hashFun = hash.HashWithSeed{Seed: seed}

	h.buckets = make([]uint8, h.m)
	for i := 0; i < int(h.m); i++ {
		n, err = file.Read(buffer1[:])
		if err != nil {
			fmt.Println("failed to read bucket", err)
			return err
		}
		if n != 1 {
			return fmt.Errorf("read only %d, expected 1", n)
		}
		h.buckets[i] = serialization.BytesToUInt8(buffer1[:])
	}

	return nil
}
