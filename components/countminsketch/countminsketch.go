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
		fmt.Println("faild to read bytes ", err)
		return err
	}
	if n != len(buffer4) {
		fmt.Printf("read only %d, but expected %d\n", n, len(buffer4))
		return fmt.Errorf("read only %d, but expected %d", n, len(buffer4))
	}
	cms.k = serialization.BytesToUInt32(buffer4[:])

	n, err = file.Read(buffer4[:])
	if err != nil {
		fmt.Println("faild to read bytes ", err)
		return err
	}
	if n != len(buffer4) {
		fmt.Printf("read only %d, but expected %d\n", n, len(buffer4))
		return fmt.Errorf("read only %d, but expected %d", n, len(buffer4))
	}
	cms.m = serialization.BytesToUInt32(buffer4[:])

	// Reset hashFun
	cms.hashFun = make([]hash.HashWithSeed, 0, cms.k)

	for i := 0; i < int(cms.k); i++ {
		h := hash.HashWithSeed{
			Seed: make([]byte, 4),
		}
		n, err = file.Read(h.Seed)
		if err != nil {
			fmt.Println("faild to read bytes ", err)
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
				fmt.Println("faild to read bytes ", err)
				return err
			}
			if n != len(buffer8) {
				fmt.Printf("read only %d, but expected %d\n", n, len(buffer8))
				return fmt.Errorf("read only %d, but expected %d", n, len(buffer8))
			}
			v := serialization.BytesToUInt64(buffer8[:])
			row[j] = v
		}
		cms.matrix[i] = row
	}
	return nil
}
