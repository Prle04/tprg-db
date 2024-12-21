package simhash

import (
	"encoding/binary"
	"regexp"
	"strings"
	"tprgdb/components/hash"
)

// reci koje se ne posmatraju, jer nemaju nikakvu vrednost
var bannedWords = map[string]int{
	"a":   0,
	"an":  0,
	"the": 0,
	"but": 0,
	"if":  0,
	"or":  0,
}

type Simhash struct {
	fingerprint uint64
}

// za svaku rec broj ponavljanja i njen hash
type wordData struct {
	numberOfOccurences uint32
	hash               uint64
}

// izbacivanje znakova interpunkcije i prebacivnje svih slova u mala
func transformText(s string) string {
	r := regexp.MustCompile(`[[:punct:]]`)
	s = r.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

func splitText(s string) []string {
	r := regexp.MustCompile(`\s+`)
	return r.Split(s, -1)
}

// mapa koja ce sadrzati reci kao kljuc, a vrednost ce biti tipa wordData
func createMap(s string) map[string]wordData {

	result := splitText(transformText(s))

	m := make(map[string]wordData)

	for _, res := range result {

		_, yes := bannedWords[res]

		if yes {
			continue
		}

		b := []byte(res)

		val, ok := m[res]
		if !ok {
			val = wordData{
				numberOfOccurences: 0,
				hash:               hash.GetMD5Hash(b),
			}
		}
		val.numberOfOccurences++
		m[res] = val

	}

	return m
}

// sabranje bitova na istoj poziciji po zadatoj semi
func sumBitsFromMap(m map[string]wordData) [64]int {

	var sumBit [64]int

	for _, value := range m {
		h := value.hash
		count := value.numberOfOccurences

		for i := 0; i < 64; i++ {
			bit := (h >> uint(63-i)) & 1

			if bit == 1 {
				sumBit[i] += int(count)
			} else {
				sumBit[i] -= int(count)
			}

		}
	}
	return sumBit
}

// krajnji simhash
func calculate(m map[string]wordData) uint64 {

	sumBit := sumBitsFromMap(m)

	var simhash uint64
	for i := 0; i < 64; i++ {
		if sumBit[i] > 0 {
			simhash |= 1 << uint(63-i)
		}
	}

	return simhash
}

func Create(s string) Simhash {
	m := createMap(s)

	fp := Simhash{
		fingerprint: calculate(m),
	}

	return fp
}

func HammingDistance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2

	distance := 0
	for xor != 0 {
		if xor&1 == 1 {
			distance++
		}
		xor >>= 1
	}

	return distance
}

func (s Simhash) Serialize() []byte {
	byteArray := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteArray, s.fingerprint)
	return byteArray
}
