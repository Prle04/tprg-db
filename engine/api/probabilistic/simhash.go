package probabilistic

import (
	"errors"
	"fmt"
	"tprgdb/components/serialization"
	"tprgdb/components/simhash"
	"tprgdb/engine/constants"
	"tprgdb/engine/core"
)

type SimHashAPI struct {
	core *core.Core
}

func NewSimHashAPI(core *core.Core) *SimHashAPI {
	return &SimHashAPI{
		core: core,
	}
}
func (api *SimHashAPI) Create(name string, text string) error {
	if name == "" {
		return errors.New("fingerprint name cannot be empty")
	}
	sh := simhash.SimHash(text)
	bytes := serialization.UInt64ToBytes(sh)
	return api.core.Set(constants.SH_PREFIX+name, bytes)
}

func (api *SimHashAPI) Delete(name string) error {
	if name == "" {
		return errors.New("fingerprint name cannot be empty")
	}
	return api.core.Delete(constants.SH_PREFIX + name)
}

func (api *SimHashAPI) Get(name string) (uint64, error) {
	if name == "" {
		return 0, errors.New("fingerprint name cannot be empty")
	}

	item, err := api.core.Get(constants.SH_PREFIX + name)
	if err != nil {
		return 0, err
	}
	if item == nil || item.IsDeleted() {
		return 0, errors.New("fingerprint does not exist")
	}

	return serialization.BytesToUInt64(item.GetValue()), nil
}

func (api *SimHashAPI) Similarity(name1, name2 string) (int, error) {
	if name1 == "" || name2 == "" {
		return 0, errors.New("fingerprint names cannot be empty")
	}

	fingerprints, err := api.core.GetMany([]string{constants.SH_PREFIX + name1, constants.SH_PREFIX + name2})
	if err != nil {
		return 0, err
	}
	if len(fingerprints) != 2 {
		return 0, errors.New("both fingerprints must exist")
	}
	if fingerprints[0] == nil || fingerprints[0].IsDeleted() {
		return 0, fmt.Errorf("%s fingerprint does not exist", name1)
	}
	if fingerprints[1] == nil || fingerprints[1].IsDeleted() {
		return 0, fmt.Errorf("%s fingerprint does not exist", name2)
	}
	fg1 := serialization.BytesToUInt64(fingerprints[0].GetValue())
	fg2 := serialization.BytesToUInt64(fingerprints[1].GetValue())

	return simhash.HammingDistance(fg1, fg2), nil
}
