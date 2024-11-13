package bloomfilter

import "math"

func calculateM(expectedElements int, falsePositiveRate float64) uint {
	return uint(math.Ceil(float64(expectedElements) * math.Abs(math.Log(falsePositiveRate)) / math.Pow(math.Log(2), float64(2))))
}

func calculateK(expectedElements int, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(expectedElements)) * math.Log(2)))
}

func calculateNumberOfElements(m uint64) uint64 {
	return uint64(math.Ceil(float64(float64(m) / float64(ELEMENT_SIZE))))
}
