package countminsketch

import (
	"fmt"
	"io"
	"math"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type CountMinSketch struct {
	k, m    uint32
	matrix  [][]uint64
	hashFun []hash.HashWithSeed
}

const CURRENT_VERSION uint8 = 1

func New(k, m uint32) CountMinSketch {
	cms := CountMinSketch{
		k: k,
		m: m,
	}
	cms.matrix = make([][]uint64, k)
	cms.hashFun = hash.CreateHashFunctions(k)

	for i := 0; i < int(k); i++ {
		cms.matrix[i] = make([]uint64, m)
	}
	return cms
}

func (cms *CountMinSketch) Store(key []byte) {
	for i := 0; i < int(cms.k); i++ {
		h := cms.hashFun[i]
		j := h.Hash(key) % uint64(cms.m)
		cms.matrix[i][j] += 1
	}
}

func (cms CountMinSketch) IsAny(key []byte) (num uint64) {
	num = math.MaxUint64
	for i := 0; i < int(cms.k); i++ {
		h := cms.hashFun[i]
		j := h.Hash(key) % uint64(cms.m)
		value := cms.matrix[i][j]
		if value < num {
			num = value
		}
	}
	return num
}

func (cms CountMinSketch) Save(file io.Writer) error {
	bytes := serialization.UInt8ToBytes(CURRENT_VERSION)
	n, err := file.Write(bytes)
	if err != nil {
		fmt.Println("faild to save current version to file", err)
		return err
	}
	if n != len(bytes) {
		fmt.Printf("saved only %d, but expected %d\n", n, len(bytes))
		return fmt.Errorf("saved only %d, but expected %d", n, len(bytes))
	}

	bytes = serialization.UInt32ToBytes(cms.k)
	n, err = file.Write(bytes)
	if err != nil {
		fmt.Println("faild to save k to file", err)
		return err
	}
	if n != len(bytes) {
		fmt.Printf("saved only %d, but expected %d\n", n, len(bytes))
		return fmt.Errorf("saved only %d, but expected %d", n, len(bytes))
	}

	bytes = serialization.UInt32ToBytes(cms.m)
	n, err = file.Write(bytes)
	if err != nil {
		fmt.Println("faild to save m to file", err)
		return err
	}
	if n != len(bytes) {
		fmt.Printf("saved only %d, but expected %d\n", n, len(bytes))
		return fmt.Errorf("saved only %d, but expected %d", n, len(bytes))
	}

	for _, h := range cms.hashFun {
		n, err = file.Write(h.Seed)
		if err != nil {
			fmt.Println("faild to save hash seed to file", err)
			return err
		}
		if n != len(h.Seed) {
			fmt.Printf("saved only %d, but expected %d\n", n, len(h.Seed))
			return fmt.Errorf("saved only %d, but expected %d", n, len(h.Seed))
		}
	}

	for _, row := range cms.matrix {
		for _, v := range row {
			bytes = serialization.UInt64ToBytes(v)
			n, err = file.Write(bytes)
			if err != nil {
				fmt.Println("faild to save value to file", err)
				return err
			}
			if n != len(bytes) {
				fmt.Printf("saved only %d, but expected %d\n", n, len(bytes))
				return fmt.Errorf("saved only %d, but expected %d", n, len(bytes))
			}
		}
	}
	return nil
}

func (cms *CountMinSketch) Load(file io.Reader) error {
	buffer1 := [1]byte{}
	n, err := file.Read(buffer1[:])
	if err != nil {
		fmt.Println("faild to load current version", err)
		return err
	}
	if n != len(buffer1) {
		fmt.Printf("read only %d, but expected %d\n", n, len(buffer1))
		return fmt.Errorf("read only %d, but expected %d", n, len(buffer1))
	}

	version := serialization.BytesToUInt8(buffer1[:])
	if version != CURRENT_VERSION {
		fmt.Printf("got version %d, but expected %d\n", version, CURRENT_VERSION)
		return fmt.Errorf("got version %d, but expected %d", version, CURRENT_VERSION)
	}

	buffer4 := [4]byte{}
	n, err = file.Read(buffer4[:])
	if err != nil {
		fmt.Println("faild to read k", err)
		return err
	}
	if n != len(buffer4) {
		fmt.Printf("read only %d, but expected %d\n", n, len(buffer4))
		return fmt.Errorf("read only %d, but expected %d", n, len(buffer4))
	}
	cms.k = serialization.BytesToUInt32(buffer4[:])

	n, err = file.Read(buffer4[:])
	if err != nil {
		fmt.Println("faild to read m", err)
		return err
	}
	if n != len(buffer4) {
		fmt.Printf("read only %d, but expected %d\n", n, len(buffer4))
		return fmt.Errorf("read only %d, but expected %d", n, len(buffer4))
	}
	cms.m = serialization.BytesToUInt32(buffer4[:])

	cms.hashFun = make([]hash.HashWithSeed, 0, cms.k)

	for i := 0; i < int(cms.k); i++ {
		h := hash.HashWithSeed{
			Seed: make([]byte, 4),
		}
		n, err = file.Read(h.Seed)
		if err != nil {
			fmt.Println("faild to read hash seed", err)
			return err
		}
		if n != len(h.Seed) {
			fmt.Printf("read only %d, but expected %d\n", n, len(h.Seed))
			return fmt.Errorf("read only %d, but expected %d", n, len(h.Seed))
		}
		cms.hashFun = append(cms.hashFun, h)
	}

	cms.matrix = make([][]uint64, cms.k)
	buffer8 := [8]byte{}
	for i := 0; i < int(cms.k); i++ {
		row := make([]uint64, cms.m)
		for j := 0; j < int(cms.m); j++ {
			n, err = file.Read(buffer8[:])
			if err != nil {
				fmt.Println("faild to read value", err)
				return err
			}
			if n != len(buffer8) {
				fmt.Printf("read only %d, but expected %d\n", n, len(buffer8))
				return fmt.Errorf("read only %d, but expected %d", n, len(buffer8))
			}
			value := serialization.BytesToUInt64(buffer8[:])
			row[j] = value
		}
		cms.matrix[i] = row
	}
	return nil
}

func (cms CountMinSketch) SerializeToBytes() []byte {
	var result []byte

	versionBytes := serialization.UInt8ToBytes(CURRENT_VERSION)
	result = append(result, versionBytes...)

	kBytes := serialization.UInt32ToBytes(cms.k)
	result = append(result, kBytes...)

	mBytes := serialization.UInt32ToBytes(cms.m)
	result = append(result, mBytes...)

	for i := 0; i < int(cms.k); i++ {
		seed := cms.hashFun[i].Seed
		result = append(result, seed...)
	}

	for i := 0; i < int(cms.k); i++ {
		for j := 0; j < int(cms.m); j++ {
			value := cms.matrix[i][j]
			valueBytes := serialization.UInt64ToBytes(value)
			result = append(result, valueBytes...)
		}
	}

	return result
}

func (cms *CountMinSketch) DeserializeFromBytes(data []byte) error {
	offset := 0

	if len(data) < 9 {
		return fmt.Errorf("data too short to deserialize")
	}

	version := serialization.BytesToUInt8(data[offset : offset+1])
	if version != CURRENT_VERSION {
		return fmt.Errorf("got version %d, but expected %d", version, CURRENT_VERSION)
	}
	offset += 1

	cms.k = serialization.BytesToUInt32(data[offset : offset+4])
	offset += 4

	cms.m = serialization.BytesToUInt32(data[offset : offset+4])
	offset += 4

	cms.hashFun = make([]hash.HashWithSeed, cms.k)
	for i := 0; i < int(cms.k); i++ {
		if offset+4 > len(data) {
			return fmt.Errorf("not enough data for hash seed %d", i)
		}
		seed := make([]byte, 4)
		copy(seed, data[offset:offset+4])
		cms.hashFun[i] = hash.HashWithSeed{
			Seed: seed,
		}
		offset += 4
	}

	cms.matrix = make([][]uint64, cms.k)
	for i := 0; i < int(cms.k); i++ {
		row := make([]uint64, cms.m)
		for j := 0; j < int(cms.m); j++ {
			if offset+8 > len(data) {
				return fmt.Errorf("not enough data for matrix[%d][%d]", i, j)
			}
			value := serialization.BytesToUInt64(data[offset : offset+8])
			offset += 8
			row[j] = value
		}
		cms.matrix[i] = row
	}

	return nil
}
