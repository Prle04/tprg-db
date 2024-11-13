package bloomfilter

import (
	"errors"
	"os"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

// implements Serializable interface
type BloomFilter struct {
	m      uint64
	k      uint64
	hashes []hash.HashWithSeed
	data   []uint8
}

const ELEMENT_SIZE uint8 = 8
const CURRENT_VERSION uint8 = 1

func NewBloomFilter(expectedElements int, falsePositiveRate float64) BloomFilter {
	var bloomfilter BloomFilter
	bloomfilter.m = uint64(calculateM(expectedElements, falsePositiveRate))
	bloomfilter.k = uint64(calculateK(expectedElements, uint(bloomfilter.m)))
	bloomfilter.hashes = hash.CreateHashFunctions(uint32(bloomfilter.k))
	var numberOfElements uint64 = calculateNumberOfElements(bloomfilter.m)
	bloomfilter.m = numberOfElements * uint64(ELEMENT_SIZE)
	bloomfilter.data = make([]uint8, numberOfElements)
	return bloomfilter
}

func (bl *BloomFilter) Insert(data []byte) {
	for i := uint64(0); i < bl.k; i++ {
		var h uint64 = bl.hashes[i].Hash(data)
		elementIndex, elementBit := bl.calculateIndecies(h)
		var mask uint8 = (1 << elementBit)
		bl.data[elementIndex] |= mask
	}
}

func (bl *BloomFilter) IsNotIn(data []byte) bool {
	for i := uint64(0); i < bl.k; i++ {
		var h uint64 = bl.hashes[i].Hash(data)
		elementIndex, elementBit := bl.calculateIndecies(h)
		var mask uint8 = (1 << elementBit)
		if bl.data[elementIndex]&mask == 0 {
			return true
		}
	}
	return false
}

func (bl *BloomFilter) calculateIndecies(hash uint64) (uint64, uint8) {
	var index uint64 = hash % bl.m
	var elementIndex uint64 = index / uint64(ELEMENT_SIZE)
	var elementBit uint8 = uint8(index % uint64(ELEMENT_SIZE))

	return elementIndex, elementBit
}

func (bl *BloomFilter) SerializeToFile(file os.File) error {
	// 1B - Version
	// 8B - m
	// 8B - k
	// k*4B - hash seeds
	// nofEl * 1B - data

	// Version
	_, err := file.Write(serialization.UInt8ToBytes(CURRENT_VERSION))
	if err != nil {
		return err
	}
	// m
	_, err = file.Write(serialization.UInt64ToBytes(bl.m))
	if err != nil {
		return err
	}
	// k
	_, err = file.Write(serialization.UInt64ToBytes(bl.k))
	if err != nil {
		return err
	}	
	// k hash seeds
	for i := uint64(0); i < bl.k; i++ {
		_, err = file.Write(bl.hashes[i].Seed)
		if err != nil {
			return err
		}
	}
	// elements
	var numberOfElements uint64 = calculateNumberOfElements(bl.m)
	for i := uint64(0); i < numberOfElements; i++ {
		_, err = file.Write(serialization.UInt8ToBytes(bl.data[i]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (bl *BloomFilter) DeserializeFromFile(file os.File) error {
	// 1B - Version
	// 8B - m
	// 8B - k
	// k*4B - hash seeds
	// nofEl * 1B - data

	// buffer for reading
	var buffer1 []byte = make([]byte, 1)
	// var buffer2 []byte = make([]byte, 2)
	var buffer4 []byte = make([]byte, 4)
	var buffer8 []byte = make([]byte, 8)

	n, err := file.Read(buffer1)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("error reading version")
	}
	var version uint8 = serialization.BytesToUInt8(buffer1)
	if version != CURRENT_VERSION {
		// handle diffrent version
		return errors.New("version mismatch")
	}

	// m
	n, err = file.Read(buffer8)
	if err != nil {
		return err
	}
	if n != 8 {
		return errors.New("error reading m")
	}
	bl.m = serialization.BytesToUInt64(buffer8)

	// k
	n, err = file.Read(buffer8)
	if err != nil {
		return err
	}
	if n != 8 {
		return errors.New("error reading k")
	}
	bl.k = serialization.BytesToUInt64(buffer8)

	// seeds
	bl.hashes = make([]hash.HashWithSeed, bl.k)
	for i := uint64(0); i < bl.k; i++ {
		n, err = file.Read(buffer4)
		if err != nil {
			return err
		}
		if n != 4 {
			return errors.New("error reading seed")
		}
		bl.hashes[i] = hash.HashWithSeed{}
		copy(bl.hashes[i].Seed, buffer4)	// copy bytes from buffer to seed
	}
	
	// data
	var numberOfElements uint64 = calculateNumberOfElements(bl.m)
	bl.data = make([]uint8, numberOfElements)
	for i := uint64(0); i < numberOfElements; i++ {
		n, err = file.Read(buffer1)
		if err != nil {
			return err
		}
		if n != 1 {
			return errors.New("error reading data")
		}
		bl.data[i] = serialization.BytesToUInt8(buffer1)
	}
	return nil // no error
}

func (bl *BloomFilter) SerializeToBytes() []byte {
	// 1B - Version
	// 8B - m
	// 8B - k
	// k*4B - hash seeds
	// nofEl * 1B - data

	var result []byte = []byte{}

	// Version
	result = append(result, serialization.UInt8ToBytes(CURRENT_VERSION)...)

	// m
	result = append(result, serialization.UInt64ToBytes(bl.m)...)

	// k
	result = append(result, serialization.UInt64ToBytes(bl.k)...)

	// k hash seeds
	for i := uint64(0); i < bl.k; i++ {
		result = append(result, bl.hashes[i].Seed...)
	}

	// elements
	var numberOfElements uint64 = calculateNumberOfElements(bl.m)
	for i := uint64(0); i < numberOfElements; i++ {
		result = append(result, serialization.UInt8ToBytes(bl.data[i])...)
	}

	return result
}

func (bl *BloomFilter) DeserializeFromBytes(data []byte) {
	// 1B - Version
	// 8B - m
	// 8B - k
	// k*4B - hash seeds
	// nofEl * 1B - data

	var version uint8 = uint8(data[0])
	if version != CURRENT_VERSION {
		// handle diffrent version
		return
	}
	data = data[1:] // skip version

	// m
	bl.m = serialization.BytesToUInt64(data[:8])
	data = data[8:] // remove m
	// k
	bl.k = serialization.BytesToUInt64(data[:8])
	data = data[8:] // remove k

	// number of elements
	var numberOfElements uint64 = calculateNumberOfElements(bl.m)

	// seeds
	bl.hashes = make([]hash.HashWithSeed, bl.k)
	for i := uint64(0); i < bl.k; i++ {
		var seedBytes []byte = data[:4]
		bl.hashes[i] = hash.HashWithSeed{Seed: seedBytes}
		data = data[4:] // remove last seed
	}
	// data
	bl.data = make([]uint8, numberOfElements)
	for i := uint64(0); i < numberOfElements; i++ {
		bl.data[i] = serialization.BytesToUInt8(data[:1])
		data = data[1:] // remove last element
	}
}
