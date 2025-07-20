package probabilistic

import "tprgdb/engine/core"

type ProbabilisticAPI struct {
	BloomFilter    *BloomFilterAPI
	CountMinSketch *CountMinSketchAPI
	HyperLogLog    *HyperLogLogAPI
	SimHash        *SimHashAPI
}

func NewProbabilisticAPI(core *core.Core) *ProbabilisticAPI {
	return &ProbabilisticAPI{
		BloomFilter:    NewBloomFilterAPI(core),
		CountMinSketch: NewCountMinSketchAPI(core),
		HyperLogLog:    NewHyperLogLogAPI(core),
		SimHash:        NewSimHashAPI(core),
	}
}
