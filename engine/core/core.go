package core

import (
	"fmt"
	"time"
	"tprgdb/components/avltree"
	"tprgdb/components/blockmanager"
	"tprgdb/components/compression"
	"tprgdb/components/lrucache"
	"tprgdb/components/lsm"
	"tprgdb/components/memtable"
	"tprgdb/components/sstable"
	walPackage "tprgdb/components/wal"
	externalconfiguration "tprgdb/engine/externalConfiguration"
)

type Core struct {
	wal                *walPackage.WriteAheadLog
	blockManager       *blockmanager.BlockManager
	lsm                *lsm.LSM
	compressionManager *compression.CompressionManager
	// Insert once the AVL tree is implemented
	avl            *avltree.AVLTree
	sstableManager *sstable.SSTableManager
	memtable       *memtable.Memtable
	cache          *lrucache.LRUCache[*Item]
}

func NewCore(config *externalconfiguration.ExternalConfiguration) (*Core, error) {

	blockSize := uint64(config.BlockManager.BlockSizeBytes)

	blockManager := blockmanager.NewBlockManager(blockSize, uint64(config.BlockManager.MaxElements))

	paddingThreshold := uint64(0.1 * float64(blockSize))
	wal, err := walPackage.NewWriteAheadLog(config.WAL.WALFolder, uint64(config.WAL.NumberOfBlocksPerSegment), paddingThreshold, blockManager)
	if err != nil {
		return nil, err
	}
	avl := avltree.CreateAVLTree()
	// TODO: Use avl as compression structure once serialization is implemented
	compressionStruct := compression.NewAVLCompression(avl)
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

	for it := range wal.Iterate() {
		if it.Err != nil {
			return nil, it.Err
		}
		if it.Record == nil {
			return nil, fmt.Errorf("nil record found in WAL")
		}
		if it.Record.GetTombstone() {
			if err := memtable.Delete(it.Record.GetKey(), it.SegmentNumber); err != nil {
				return nil, fmt.Errorf("error setting deleted item in memtable: %w", err)
			}
		} else {
			if err := memtable.Set(it.Record.GetKey(), it.Record.GetValue(), it.SegmentNumber); err != nil {
				return nil, fmt.Errorf("error setting item in memtable: %w", err)
			}
		}
	}

	return &Core{
		wal:                wal,
		blockManager:       blockManager,
		lsm:                lsm,
		compressionManager: compressionManager,
		sstableManager:     sstableManager,
		memtable:           memtable,
		cache:              cache,
		avl:                avl,
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
func (c *Core) Set(key string, value []byte) error {

	compressedKey := c.compressionManager.Add(key)

	if err := c.wal.Add(key, value); err != nil {
		return err
	}
	if err := c.memtable.Set(key, value, c.wal.GetCurrentSegmentNumber()); err != nil {
		return err
	}

	c.updateCache(key, value, uint64(time.Now().Unix()), false)

	if err := c.flush(compressedKey); err != nil {
		return err
	}

	return nil
}
func (c *Core) Delete(key string) error {

	compressedKey := c.compressionManager.Add(key)
	if err := c.wal.Delete(key); err != nil {
		return err
	}
	if err := c.memtable.Delete(key, c.wal.GetCurrentSegmentNumber()); err != nil {
		return err
	}

	c.updateCache(key, nil, uint64(time.Now().Unix()), true)

	if err := c.flush(compressedKey); err != nil {
		return err
	}

	return nil
}

func (c *Core) GetMany(keys []string) ([]*Item, error) {
	items := make([]*Item, len(keys))
	nextKeys := make([]string, 0, len(keys))
	indexes := make(map[string]int, len(keys))
	for i, key := range keys {
		indexes[key] = i
		memtableItem := c.memtable.Get(key)
		if memtableItem != nil {
			items[i] = &Item{
				key:       memtableItem.GetKey(),
				value:     memtableItem.GetValue(),
				timestamp: memtableItem.GetTimestamp(),
				tombstone: memtableItem.IsDeleted(),
			}
		} else {
			cacheItem, found := c.cache.GetElement(key)
			if found {
				items[i] = &Item{
					key:       cacheItem.GetKey(),
					value:     cacheItem.GetValue(),
					timestamp: cacheItem.GetTimestamp(),
					tombstone: cacheItem.IsDeleted(),
				}
			} else {
				nextKeys = append(nextKeys, key)
			}
		}
	}
	if len(nextKeys) == 0 {
		return items, nil
	}
	sstableItems, err := c.lsm.GetMany(nextKeys)
	if err != nil {
		return nil, err
	}
	for _, sstableItem := range sstableItems {
		if sstableItem == nil {
			continue
		}
		index, exists := indexes[sstableItem.GetKey()]
		if !exists {
			continue
		}
		item := &Item{
			key:       sstableItem.GetKey(),
			value:     sstableItem.GetValue(),
			timestamp: sstableItem.GetTimestamp(),
			tombstone: sstableItem.IsDeleted(),
		}
		items[index] = item
		c.cache.AddElement(item.GetKey(), item)
	}
	return items, nil
}

func (c *Core) GetKeysRange(start, end string, pageNumber int, itemsPerPage int) ([]string, error) {
	data := c.avl.RangeScan(start, end, pageNumber, itemsPerPage)
	if data == nil {
		return nil, nil
	}
	keys := make([]string, len(data))
	for i, item := range data {
		if item == nil {
			continue
		}
		keys[i] = item.GetKey()
	}
	return keys, nil
}

func (c *Core) ValidateSSTable(tableName string) (bool, error) {
	return c.lsm.ValidateSSTable(tableName)
}

func (c *Core) updateCache(key string, value []byte, timestamp uint64, tombstone bool) {
	if c.cache.Exists(key) {
		item := &Item{
			key:       key,
			value:     value,
			timestamp: timestamp,
			tombstone: tombstone,
		}
		if added, _ := c.cache.AddElement(key, item); !added {
			// If item was not updated in the cache, it means we have a stale
			// item in the cache, so we need to remove it
			c.cache.RemoveElement(key)
		}
	}
}

func (c *Core) flush(compressedKey uint64) error {
	for items, lastSegment := c.memtable.Flush(); items != nil; items, lastSegment = c.memtable.Flush() {
		data := make([]sstable.ISSTableItem, len(items))
		for i, item := range items {
			data[i] = &sstable.SSItem{
				Key:           item.GetKey(),
				CompressedKey: compressedKey,
				Value:         item.GetValue(),
				Timestamp:     item.GetTimestamp(),
				Tombstone:     item.IsDeleted(),
			}
		}
		c.lsm.AddTable(data)
		err := c.lsm.Save()
		if err != nil {
			return err
		}
		c.wal.DeleteUntil(lastSegment)
	}
	return nil
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

func Shutdown(c *Core) error {
	if err := c.wal.Close(); err != nil {
		return fmt.Errorf("error closing WAL: %w", err)
	}
	if err := c.blockManager.Close(); err != nil {
		return fmt.Errorf("error closing BlockManager: %w", err)
	}
	if err := c.lsm.Save(); err != nil {
		return fmt.Errorf("error closing LSM: %w", err)
	}
	return nil
}
