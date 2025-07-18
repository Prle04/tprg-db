package externalconfiguration

import (
	"encoding/json"
	"os"
	"strings"
)

type ExternalConfiguration struct {
	WAL          WALConfig          `json:"walConfig"`
	Memtable     MemtableConfig     `json:"memtableConfig"`
	Cache        CacheConfig        `json:"cacheConfig"`
	BlockManager BlockManagerConfig `json:"blockManagerConfig"`
	LSM          LSMConfig          `json:"lsmConfig"`
	SSTable      SSTableConfig      `json:"sstableConfig"`
	Compression  CompressionConfig  `json:"compressionConfig"`
}

func (ec *ExternalConfiguration) Validate() {
	ec.WAL.Validate()
	ec.Memtable.Validate()
	ec.Cache.Validate()
	ec.BlockManager.Validate()
	ec.LSM.Validate()
	ec.SSTable.Validate()
	ec.Compression.Validate()
}

func NewExternalConfiguration(filepath string) *ExternalConfiguration {
	var config *ExternalConfiguration
	if !strings.HasSuffix(filepath, ".json") {
		config = &ExternalConfiguration{}
	} else {
		data, err := os.ReadFile(filepath)
		if err != nil {
			config = &ExternalConfiguration{}
		} else if err := json.Unmarshal(data, &config); err != nil {
			config = &ExternalConfiguration{}
		}
	}
	config.Validate()
	return config
}
