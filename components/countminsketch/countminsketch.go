package countminsketch

import (
	"errors"
	"os"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type CountMinSketch struct {
	width  uint64
	depth  uint64
	hashes []hash.HashWithSeed
	matrix [][]uint64
}

const CURRENT_VERSION uint8 = 1

// Kreiranje novog CountMinSketch objekta
func NewCountMinSketch(width, depth uint64) CountMinSketch {
	cms := CountMinSketch{
		width:  width,
		depth:  depth,
		hashes: hash.CreateHashFunctions(uint32(depth)),
		matrix: make([][]uint64, depth),
	}
	for i := range cms.matrix {
		cms.matrix[i] = make([]uint64, width)
	}
	return cms
}

// Ubacivanje podataka u CountMinSketch
func (cms *CountMinSketch) Insert(data []byte) {
	for i, hfn := range cms.hashes {
		hashValue := hfn.Hash(data) % cms.width
		cms.matrix[i][hashValue]++
	}
}

// Estimacija broja pojavljivanja podataka u CountMinSketch
func (cms *CountMinSketch) Estimate(data []byte) uint64 {
	minCount := uint64(^uint64(0))
	for i, hfn := range cms.hashes {
		hashValue := hfn.Hash(data) % cms.width
		if cms.matrix[i][hashValue] < minCount {
			minCount = cms.matrix[i][hashValue]
		}
	}
	return minCount
}

// SerializeToFile čuva CountMinSketch u fajl
func (cms *CountMinSketch) SerializeToFile(file os.File) error {
	_, err := file.Write(serialization.UInt8ToBytes(CURRENT_VERSION))
	if err != nil {
		return err
	}

	_, err = file.Write(serialization.UInt64ToBytes(cms.width))
	if err != nil {
		return err
	}

	_, err = file.Write(serialization.UInt64ToBytes(cms.depth))
	if err != nil {
		return err
	}

	for _, hfn := range cms.hashes {
		_, err = file.Write(hfn.Seed)
		if err != nil {
			return err
		}
	}

	for _, row := range cms.matrix {
		for _, value := range row {
			_, err = file.Write(serialization.UInt64ToBytes(value))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeserializeFromFile učitava CountMinSketch iz fajla
func (cms *CountMinSketch) DeserializeFromFile(file os.File) error {
	var buffer1 []byte = make([]byte, 1)
	var buffer8 []byte = make([]byte, 8)

	n, err := file.Read(buffer1)
	if err != nil || n != 1 {
		return errors.New("error reading version")
	}

	version := serialization.BytesToUInt8(buffer1)
	if version != CURRENT_VERSION {
		return errors.New("version mismatch")
	}

	n, err = file.Read(buffer8)
	if err != nil || n != 8 {
		return errors.New("error reading width")
	}
	cms.width = serialization.BytesToUInt64(buffer8)

	n, err = file.Read(buffer8)
	if err != nil || n != 8 {
		return errors.New("error reading depth")
	}
	cms.depth = serialization.BytesToUInt64(buffer8)

	cms.hashes = make([]hash.HashWithSeed, cms.depth)
	for i := range cms.hashes {
		n, err = file.Read(buffer8)
		if err != nil || n != 8 {
			return errors.New("error reading hash seed")
		}
		cms.hashes[i] = hash.HashWithSeed{Seed: buffer8}
	}

	cms.matrix = make([][]uint64, cms.depth)
	for i := range cms.matrix {
		cms.matrix[i] = make([]uint64, cms.width)
		for j := range cms.matrix[i] {
			n, err = file.Read(buffer8)
			if err != nil || n != 8 {
				return errors.New("error reading matrix value")
			}
			cms.matrix[i][j] = serialization.BytesToUInt64(buffer8)
		}
	}

	return nil
}

// SerializeToBytes serializuje CountMinSketch u bajt niz
func (cms *CountMinSketch) SerializeToBytes() []byte {
	var result []byte

	result = append(result, serialization.UInt8ToBytes(CURRENT_VERSION)...)
	result = append(result, serialization.UInt64ToBytes(cms.width)...)
	result = append(result, serialization.UInt64ToBytes(cms.depth)...)

	for _, hfn := range cms.hashes {
		result = append(result, hfn.Seed...)
	}

	for _, row := range cms.matrix {
		for _, value := range row {
			result = append(result, serialization.UInt64ToBytes(value)...)
		}
	}

	return result
}

// DeserializeFromBytes deserijalizuje CountMinSketch iz bajt niza
func (cms *CountMinSketch) DeserializeFromBytes(data []byte) {
	version := uint8(data[0])
	if version != CURRENT_VERSION {
		return
	}
	data = data[1:]

	cms.width = serialization.BytesToUInt64(data[:8])
	data = data[8:]

	cms.depth = serialization.BytesToUInt64(data[:8])
	data = data[8:]

	cms.hashes = make([]hash.HashWithSeed, cms.depth)
	for i := range cms.hashes {
		cms.hashes[i] = hash.HashWithSeed{Seed: data[:8]}
		data = data[8:]
	}

	cms.matrix = make([][]uint64, cms.depth)
	for i := range cms.matrix {
		cms.matrix[i] = make([]uint64, cms.width)
		for j := range cms.matrix[i] {
			cms.matrix[i][j] = serialization.BytesToUInt64(data[:8])
			data = data[8:]
		}
	}
}
