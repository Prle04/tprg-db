package externalconfiguration

type ExternalConfiguration struct {
	WAL          WALConfig                `json:"wal"`
	Memtable     MemtableConfig           `json:"memtable"`
	Cache        CacheConfig              `json:"cache"`
	BlockManager BlockManagerConfig       `json:"blockManager"`
	LSM          LSMConfig                `json:"lsm"`
	SSTable      SSTableConfig            `json:"sstable"`
	Compression  CompressionConfig        `json:"compression"`
	TokenBucket  TokenBucketConfiguration `json:"tokenBucket"`
}

func (ec *ExternalConfiguration) Validate(dbName string) {
	ec.WAL.Validate(dbName)
	ec.Memtable.Validate(dbName)
	ec.Cache.Validate(dbName)
	ec.BlockManager.Validate(dbName)
	ec.LSM.Validate(dbName)
	ec.SSTable.Validate(dbName)
	ec.Compression.Validate(dbName)
	ec.TokenBucket.Validate(dbName)
}
