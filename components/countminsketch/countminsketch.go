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
	minCount := uint64(^uint64(0)) // Maksimalna vrednost uint64
	for i, hfn := range cms.hashes {
		hashValue := hfn.Hash(data) % cms.width
		if cms.matrix[i][hashValue] < minCount {
			minCount = cms.matrix[i][hashValue]
		}
	}
	return minCount
}

// Serijalizacija u bajtove
func (cms *CountMinSketch) SerializeToBytes() []byte {
	result := []byte{}
	result = append(result, serialization.UInt8ToBytes(CURRENT_VERSION)...)
	result = append(result, serialization.UInt64ToBytes(cms.width)...)
	result = append(result, serialization.UInt64ToBytes(cms.depth)...)

	// Dodaj hash seedove
	for _, h := range cms.hashes {
		result = append(result, h.Seed...)
	}

	// Dodaj matricu
	for i := range cms.matrix {
		for _, count := range cms.matrix[i] {
			result = append(result, serialization.UInt64ToBytes(count)...)
		}
	}
	return result
}

// Deserijalizacija iz bajtova
func (cms *CountMinSketch) DeserializeFromBytes(data []byte) error {
	version := serialization.BytesToUInt8(data[:1])
	if version != CURRENT_VERSION {
		return errors.New("version mismatch")
	}
	data = data[1:]

	cms.width = serialization.BytesToUInt64(data[:8])
	data = data[8:]
	cms.depth = serialization.BytesToUInt64(data[:8])
	data = data[8:]

	cms.hashes = make([]hash.HashWithSeed, cms.depth)
	for i := range cms.hashes {
		cms.hashes[i] = hash.HashWithSeed{Seed: data[:32]}
		data = data[32:]
	}

	cms.matrix = make([][]uint64, cms.depth)
	for i := range cms.matrix {
		cms.matrix[i] = make([]uint64, cms.width)
		for j := range cms.matrix[i] {
			cms.matrix[i][j] = serialization.BytesToUInt64(data[:8])
			data = data[8:]
		}
	}
	return nil
}

// Serijalizacija u fajl
func (cms *CountMinSketch) SerializeToFile(file *os.File) error {
	_, err := file.Write(cms.SerializeToBytes())
	return err
}

// Deserijalizacija iz fajla
func (cms *CountMinSketch) DeserializeFromFile(file *os.File) error {
	data := make([]byte, fileSize(file))
	_, err := file.Read(data)
	if err != nil {
		return err
	}
	return cms.DeserializeFromBytes(data)
}

// Funkcija za dobijanje veličine fajla
func fileSize(file *os.File) int64 {
	info, _ := file.Stat()
	return info.Size()
}
