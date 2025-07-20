package probabilistic

import (
	"errors"
	"tprgdb/components/countminsketch"
	"tprgdb/engine/core"
)

const PREFIX_CMS = "C"

type CountMinSketchAPI struct {
	core *core.Core
}

func NewCountMinSketchAPI(core *core.Core) *CountMinSketchAPI {
	return &CountMinSketchAPI{
		core: core,
	}
}

func (api *CountMinSketchAPI) Create(name string, k, m int) error {
	if name == "" {
		return errors.New("cms name cannot be empty")
	}

	cms := countminsketch.New(uint32(k), uint32(m))
	bytes := cms.SerializeToBytes()
	return api.core.Set(PREFIX_CMS+name, bytes)
}

func (api *CountMinSketchAPI) Delete(name string) error {
	if name == "" {
		return errors.New("cms name cannot be empty")
	}
	return api.core.Delete(PREFIX_CMS + name)
}

func (api *CountMinSketchAPI) Insert(name string, data []byte) error {
	if name == "" {
		return errors.New("cms name cannot be empty")
	}
	if len(data) == 0 {
		return errors.New("data to insert cannot be empty")
	}

	item, err := api.core.Get(PREFIX_CMS + name)
	if err != nil {
		return err
	}
	if item == nil || item.IsDeleted() {
		return errors.New("cms does not exist")
	}

	cms := &countminsketch.CountMinSketch{}

	cms.DeserializeFromBytes(item.GetValue())

	cms.Store(data)
	return api.core.Set(PREFIX_CMS+name, cms.SerializeToBytes())
}

func (api *CountMinSketchAPI) Count(name string, data []byte) (uint64, error) {
	if name == "" {
		return 0, errors.New("cms name cannot be empty")
	}
	if len(data) == 0 {
		return 0, errors.New("data to count cannot be empty")
	}

	item, err := api.core.Get(PREFIX_CMS + name)
	if err != nil {
		return 0, err
	}
	if item == nil || item.IsDeleted() {
		return 0, errors.New("cms does not exist")
	}

	cms := &countminsketch.CountMinSketch{}
	cms.DeserializeFromBytes(item.GetValue())

	return cms.IsAny(data), nil
}
