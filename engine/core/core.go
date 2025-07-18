package core

import (
	"tprgdb/components/blockmanager"
	"tprgdb/components/compression"
	"tprgdb/components/lrucache"
	"tprgdb/components/lsm"
	"tprgdb/components/memtable"
	"tprgdb/components/sstable"
	externalconfiguration "tprgdb/engine/externalConfiguration"
)

type Core struct {
	blockManager       *blockmanager.BlockManager
	lsm                *lsm.LSM
	compressionManager *compression.CompressionManager
	// Insert once the AVL tree is implemented
	// avl                *avltree.AVLTree
	sstableManager *sstable.SSTableManager
	memtable       *memtable.Memtable
	cache          *lrucache.LRUCache[*Item]
}

func NewCore(config *externalconfiguration.ExternalConfiguration) (*Core, error) {

	blockManager := blockmanager.NewBlockManager(uint64(config.BlockManager.BlockSizeBytes), uint64(config.BlockManager.MaxElements))

	compressionStruct := compression.NewHashMapCompression()
	decompressionStruct := compression.NewHashMapDecompression()
	compressionManager := compression.NewCompressionManager(compressionStruct, decompressionStruct, config.Compression.Enabled)

	cache := lrucache.NewLRUCache[*Item](uint64(config.Cache.MaxElements))

	sstableManager := sstable.NewSSTableManager(config.SSTable.FolderPath, config.SSTable.Mode, uint8(config.SSTable.SummaryDilutionFactor), blockManager, compressionManager)

	var filesPerLevel int
	var multiplier int
	if config.LSM.Algorithm == lsm.SIZE_TIERED {
		filesPerLevel = config.LSM.SizeTier.FilesPerLevel
		multiplier = 1
	} else {
		filesPerLevel = config.LSM.Leved.InitialLevelSize
		multiplier = config.LSM.Leved.LevelMultiplier
	}
	lsm := lsm.NewLSM(config.LSM.MaxLevels, config.LSM.FolderPath, config.LSM.Algorithm, blockManager, sstableManager, compressionManager, filesPerLevel, multiplier)

	memtable, err := memtable.NewMemtable(config.Memtable.NumberOfInstances, config.Memtable.Threshld, CreateInstance(&config.Memtable))

	if err != nil {
		return nil, err
	}
	return &Core{
		blockManager:       blockManager,
		lsm:                lsm,
		compressionManager: compressionManager,
		sstableManager:     sstableManager,
		memtable:           memtable,
		cache:              cache,
	}, nil
}

func (c *Core) Get(key string) (*Item, error) {
	memtableItem := c.memtable.Get(key)
	if memtableItem != nil {
		return &Item{
			key:       memtableItem.GetKey(),
			value:     memtableItem.GetValue(),
			timestamp: memtableItem.GetTimestamp(),
			tombstone: memtableItem.IsDeleted(),
		}, nil
	}
	cacheItem, found := c.cache.GetElement(key)
	if found {
		return &Item{
			key:       cacheItem.GetKey(),
			value:     cacheItem.GetValue(),
			timestamp: cacheItem.GetTimestamp(),
			tombstone: cacheItem.IsDeleted(),
		}, nil
	}
	sstableItem, err := c.lsm.Get(key)
	if err != nil {
		return nil, err
	}
	if sstableItem != nil {
		item := &Item{
			key:       sstableItem.GetKey(),
			value:     sstableItem.GetValue(),
			timestamp: sstableItem.GetTimestamp(),
			tombstone: sstableItem.IsDeleted(),
		}
		c.cache.AddElement(item.GetKey(), item)
		return item, nil
	}
	return nil, nil
}

type Item struct {
	key       string
	value     []byte
	timestamp uint64
	tombstone bool
}

func (i Item) GetKey() string {
	return i.key
}
func (i Item) GetValue() []byte {
	return i.value
}
func (i Item) GetTimestamp() uint64 {
	return i.timestamp
}
func (i Item) IsDeleted() bool {
	return i.tombstone
}
