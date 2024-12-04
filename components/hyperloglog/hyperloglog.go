package hyperloglog

import (
	"errors"
	"fmt"
	"os"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type HyperLogLog struct {
	bits     uint64
	register []uint8
	hashes   []hash.HashWithSeed
}

const CURRENT_VERSION uint8 = 1

// Kreiranje novog HyperLogLog objekta
func NewHyperLogLog(bits uint64) HyperLogLog {
	register := make([]uint8, 1<<bits)
	hashes := hash.CreateHashFunctions(uint32(bits))
	return HyperLogLog{
		bits:     bits,
		register: register,
		hashes:   hashes,
	}
}

// Ubacivanje podataka u HyperLogLog
func (h *HyperLogLog) Insert(data []byte) {
	for _, hfn := range h.hashes {
		hashValue := hfn.Hash(data)
		index := hashValue & uint64(len(h.register)-1)
		leadingZeros := countLeadingZeros(hashValue)
		if h.register[index] < leadingZeros {
			h.register[index] = leadingZeros
		}
	}
}

/*
// Estimacija broja jedinstvenih elemenata u HyperLogLog
func (h *HyperLogLog) Estimate() float64 {
	sum := 0.0
	for _, reg := range h.register {
		sum += float64(1 << reg) // računaj sumu
	}

	alphaMM := 0.7213 / (1 + 1.079/float64(1<<int(h.bits)))
	return alphaMM * float64(len(h.register)) * float64(len(h.register)) / sum
}
*/

// Pomoćna funkcija za brojanje vodećih nula
func countLeadingZeros(value uint64) uint8 {
	return uint8(64 - len(fmt.Sprintf("%b", value)))
}

// Implementacija SerializeToBytes za HyperLogLog
func (h *HyperLogLog) SerializeToBytes() []byte {
	var result []byte
	result = append(result, serialization.UInt8ToBytes(CURRENT_VERSION)...)
	result = append(result, serialization.UInt64ToBytes(h.bits)...)
	for _, hfn := range h.hashes {
		result = append(result, hfn.Seed...)
	}
	for _, reg := range h.register {
		result = append(result, serialization.UInt8ToBytes(reg)...)
	}

	return result
}

// Implementacija DeserializeFromBytes za HyperLogLog
func (h *HyperLogLog) DeserializeFromBytes(data []byte) {
	version := uint8(data[0])
	if version != CURRENT_VERSION {
		return
	}
	data = data[1:]

	// Učitavamo broj bitova
	h.bits = serialization.BytesToUInt64(data[:8])
	data = data[8:]

	// Učitavamo hash seedove
	h.hashes = make([]hash.HashWithSeed, 1<<h.bits)
	for i := range h.hashes {
		h.hashes[i] = hash.HashWithSeed{Seed: data[:8]}
		data = data[8:]
	}

	// Učitavamo registre
	h.register = make([]uint8, 1<<h.bits)
	for i := range h.register {
		h.register[i] = serialization.BytesToUInt8(data[:1])
		data = data[1:]
	}
}

// Implementacija SerializeToFile za HyperLogLog
func (h *HyperLogLog) SerializeToFile(file os.File) error {
	_, err := file.Write(serialization.UInt8ToBytes(CURRENT_VERSION))
	if err != nil {
		return err
	}

	_, err = file.Write(serialization.UInt64ToBytes(h.bits))
	if err != nil {
		return err
	}

	for _, hfn := range h.hashes {
		_, err = file.Write(hfn.Seed)
		if err != nil {
			return err
		}
	}

	for _, reg := range h.register {
		_, err = file.Write(serialization.UInt8ToBytes(reg))
		if err != nil {
			return err
		}
	}

	return nil
}

// Implementacija DeserializeFromFile za HyperLogLog
func (h *HyperLogLog) DeserializeFromFile(file os.File) error {
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

	// Učitavamo broj bitova
	n, err = file.Read(buffer8)
	if err != nil || n != 8 {
		return errors.New("error reading bits")
	}
	h.bits = serialization.BytesToUInt64(buffer8)

	// Učitavamo hash seedove
	h.hashes = make([]hash.HashWithSeed, 1<<h.bits)
	for i := range h.hashes {
		n, err = file.Read(buffer8)
		if err != nil || n != 8 {
			return errors.New("error reading hash seed")
		}
		h.hashes[i] = hash.HashWithSeed{Seed: buffer8}
	}

	// Učitavamo registre
	h.register = make([]uint8, 1<<h.bits)
	for i := range h.register {
		n, err = file.Read(buffer1)
		if err != nil || n != 1 {
			return errors.New("error reading register")
		}
		h.register[i] = serialization.BytesToUInt8(buffer1)
	}

	return nil
}
