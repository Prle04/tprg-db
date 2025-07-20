package probabilistic

import (
	"errors"
	"tprgdb/components/hyperloglog"
	"tprgdb/engine/core"
)

const PREFIX_HLL = "H"

type HyperLogLogAPI struct {
	core *core.Core
}

func NewHyperLogLogAPI(core *core.Core) *HyperLogLogAPI {
	return &HyperLogLogAPI{
		core: core,
	}
}

func (api *HyperLogLogAPI) Create(name string, bits int) error {
	if name == "" {
		return errors.New("HLL name cannot be empty")
	}

	hll := hyperloglog.New(uint64(bits))
	bytes := hll.SerializeToBytes()
	return api.core.Set(PREFIX_HLL+name, bytes)
}

func (api *HyperLogLogAPI) Delete(name string) error {
	if name == "" {
		return errors.New("HLL name cannot be empty")
	}
	return api.core.Delete(PREFIX_HLL + name)
}

func (api *HyperLogLogAPI) Insert(name string, data []byte) error {
	if name == "" {
		return errors.New("HLL name cannot be empty")
	}
	if len(data) == 0 {
		return errors.New("data to insert cannot be empty")
	}

	item, err := api.core.Get(PREFIX_HLL + name)
	if err != nil {
		return err
	}
	if item == nil || item.IsDeleted() {
		return errors.New("HLL does not exist")
	}

	hll := &hyperloglog.HyperLogLog{}
	hll.DeserializeFromBytes(item.GetValue())

	hll.Insert(data)
	return api.core.Set(PREFIX_HLL+name, hll.SerializeToBytes())
}

func (api *HyperLogLogAPI) Count(name string) (uint64, error) {
	if name == "" {
		return 0, errors.New("HLL name cannot be empty")
	}

	item, err := api.core.Get(PREFIX_HLL + name)
	if err != nil {
		return 0, err
	}
	if item == nil || item.IsDeleted() {
		return 0, errors.New("HLL does not exist")
	}

	hll := &hyperloglog.HyperLogLog{}
	hll.DeserializeFromBytes(item.GetValue())

	return uint64(hll.Estimate()), nil
}
