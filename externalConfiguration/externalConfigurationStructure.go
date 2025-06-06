package externalconfiguration

import (
	"encoding/json"
	"os"
	"strings"
)

type ExternalConfiguration struct {
	WALConfig          WALConfig          `json:"walConfig"`
	MemtableConfig     MemtableConfig     `json:"memtableConfig"`
	CacheConfig        CacheConfig        `json:"cacheConfig"`
	BlockManagerConfig BlockManagerConfig `json:"blockManagerConfig"`
}

func (ec *ExternalConfiguration) Validate() {
	ec.WALConfig.Validate()
	ec.MemtableConfig.Validate()
	ec.CacheConfig.Validate()
	ec.BlockManagerConfig.Validate()
}

func NewExternalConfiguration(filepath string) *ExternalConfiguration {
	var config *ExternalConfiguration
	if !strings.HasSuffix(filepath, ".json") {
		config = &ExternalConfiguration{}
	} else {
		data, err := os.ReadFile(filepath)
		if err != nil {
			config = &ExternalConfiguration{}
		}
		if err := json.Unmarshal(data, &config); err != nil {
			config = &ExternalConfiguration{}
		}
	}
	config.Validate()
	return config
}
