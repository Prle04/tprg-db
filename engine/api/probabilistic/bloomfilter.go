package probabilistic

import (
	"errors"
	"tprgdb/components/bloomfilter"
	"tprgdb/engine/core"
)

const PREFIX_BLOOMFILTER = "F"

type BloomFilterAPI struct {
	core *core.Core
}

func NewBloomFilterAPI(core *core.Core) *BloomFilterAPI {
	return &BloomFilterAPI{
		core: core,
	}
}

func (api *BloomFilterAPI) Create(name string, size int, errorRate float64) error {
	if name == "" {
		return errors.New("filter name cannot be empty")
	}
	if size <= 0 || errorRate <= 0 || errorRate >= 1 {
		return errors.New("invalid parameters for Bloom filter")
	}
	filter := bloomfilter.NewBloomFilter(size, errorRate)
	bytes := filter.SerializeToBytes()
	return api.core.Set(PREFIX_BLOOMFILTER+name, bytes)
}

func (api *BloomFilterAPI) Delete(name string) error {
	if name == "" {
		return errors.New("filter name cannot be empty")
	}
	return api.core.Delete(PREFIX_BLOOMFILTER + name)
}

func (api *BloomFilterAPI) Insert(name string, data []byte) error {
	if name == "" {
		return errors.New("filter name cannot be empty")
	}
	if len(data) == 0 {
		return errors.New("data to insert cannot be empty")
	}

	item, err := api.core.Get(PREFIX_BLOOMFILTER + name)
	if err != nil {
		return err
	}
	if item == nil || item.IsDeleted() {
		return errors.New("filter does not exist")
	}

	filter := bloomfilter.NewEmptyBloomFilter()

	filter.DeserializeFromBytes(item.GetValue())

	filter.Insert(data)
	return api.core.Set(PREFIX_BLOOMFILTER+name, filter.SerializeToBytes())
}

func (api *BloomFilterAPI) IsNotIn(name string, data []byte) (bool, error) {
	if name == "" {
		return false, errors.New("filter name cannot be empty")
	}
	if len(data) == 0 {
		return false, errors.New("data to check cannot be empty")
	}

	item, err := api.core.Get(PREFIX_BLOOMFILTER + name)
	if err != nil {
		return false, err
	}
	if item == nil || item.IsDeleted() {
		return false, errors.New("filter does not exist")
	}

	filter := bloomfilter.NewEmptyBloomFilter()
	filter.DeserializeFromBytes(item.GetValue())

	return filter.IsNotIn(data), nil
}
