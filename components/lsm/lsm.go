package lsm

import (
	"encoding/binary"
	"fmt"
	"math"
	"path"
	"tprgdb/components/blockmanager"
	"tprgdb/components/sstable"
)

const (
	SIZE_TIERED = "size-tier"
	LEVELED     = "leveled"

	FILE_NAME = "lsm-tree.bin"
)

type LSM struct {
	sstableManager      *sstable.SSTableManager
	compressionManager  sstable.ICompressionManager
	blockManager        *blockmanager.BlockManager
	folderPath          string
	tables              [][]*sstable.SSTableReader
	maxLevel            int
	compactionAlgorithm string
	filePerLevel        int
	multiplier          int
}

func NewLSM(maxLevel int, folderPath string, algorithm string, blockManager *blockmanager.BlockManager, sstableManager *sstable.SSTableManager,
	compressionManager sstable.ICompressionManager, filePerLevel int, multiplier int) *LSM {
	if algorithm != SIZE_TIERED && algorithm != LEVELED {
		return nil
	}
	tables := make([][]*sstable.SSTableReader, maxLevel)
	for i := 0; i < maxLevel; i++ {
		tables[i] = make([]*sstable.SSTableReader, 0)
	}
	return &LSM{
		compactionAlgorithm: algorithm,
		maxLevel:            maxLevel,
		folderPath:          folderPath,
		tables:              tables,
		blockManager:        blockManager,
		sstableManager:      sstableManager,
		compressionManager:  compressionManager,
		filePerLevel:        filePerLevel,
		multiplier:          multiplier,
	}
}

func (lsm *LSM) AddTable(data []sstable.ISSTableItem) error {
	newTableName := generateTableName()
	iterator := sstable.NewSSListIterator(data, lsm.compressionManager)
	reader, err := lsm.sstableManager.CreateSSTable(newTableName, iterator, uint64(len(data)))
	if err != nil {
		return err
	}
	lsm.tables[0] = append(lsm.tables[0], reader)
	err = lsm.compact()
	if err != nil {
		return err
	}
	return lsm.Save()
}

func (lsm *LSM) Get(key string) (sstable.ISSTableItem, error) {
	for _, levelTables := range lsm.tables {
		for i := len(levelTables) - 1; i >= 0; i-- {
			reader := levelTables[i]
			item, err := reader.Get(key)
			if err != nil {
				return nil, err
			}
			if item != nil {
				return item, nil
			}
		}
	}
	return nil, nil
}

func (lsm *LSM) Save() error {
	filepath := path.Join(lsm.folderPath, FILE_NAME)
	file := lsm.blockManager.OpenBlockedFile(filepath)

	data := make([]byte, 0)
	varintBuf := make([]byte, 8)
	// Number of levels
	n := binary.PutUvarint(varintBuf, uint64(len(lsm.tables)))
	data = append(data, varintBuf[:n]...)
	// For each level, write the number of tables and their names
	for _, tables := range lsm.tables {
		n = binary.PutUvarint(varintBuf, uint64(len(tables)))
		data = append(data, varintBuf[:n]...)
		for _, table := range tables {
			tableName := table.GetTableName()
			n = binary.PutUvarint(varintBuf, uint64(len(tableName)))
			data = append(data, varintBuf[:n]...)
			data = append(data, []byte(tableName)...)
		}
	}
	err := file.Write(data)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (lsm *LSM) Load() error {
	filepath := path.Join(lsm.folderPath, FILE_NAME)
	file := lsm.blockManager.OpenBlockedFile(filepath)

	_, levels, err := file.ReadUVariable()
	if err != nil {
		return err
	}
	totalLevels := int(math.Max(float64(levels), float64(lsm.maxLevel)))
	lsm.tables = make([][]*sstable.SSTableReader, totalLevels)
	for i := 0; i < totalLevels; i++ {
		lsm.tables[i] = make([]*sstable.SSTableReader, 0)
	}

	for i := 0; i < totalLevels; i++ {
		// Read the number of tables in the level
		_, numTables, err := file.ReadUVariable()
		if err != nil {
			return err
		}
		for j := 0; j < int(numTables); j++ {
			// Read the table name length
			_, nameLength, err := file.ReadUVariable()
			if err != nil {
				return err
			}
			nameBytes, err := file.Read(uint64(nameLength))
			if err != nil {
				return err
			}
			tableName := string(nameBytes)
			reader := lsm.sstableManager.NewSSTableReader(tableName)
			if reader == nil {
				return fmt.Errorf("SSTable reader for table %s is nil", tableName)
			}
			lsm.tables[i] = append(lsm.tables[i], reader)
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (lsm *LSM) compact() error {
	if lsm.compactionAlgorithm == SIZE_TIERED {
		return lsm.sizeTierCompaction()
	} else {
		return lsm.leveledCompaction()
	}
}

func (l *LSM) leveledCompaction() error {
	for level := 0; level < l.maxLevel-1; level++ {
		if l.shouldCompactSized(level) {
			err := l.mergeSizeTier(level)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *LSM) sizeTierCompaction() error {
	for level := 0; level < l.maxLevel-1; level++ {
		if !l.shouldCompactSized(level) {
			break
		}
		err := l.mergeSizeTier(level)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lsm *LSM) shouldCompactSized(level int) bool {
	return len(lsm.tables[level]) > lsm.filePerLevel
}

func (lsm *LSM) mergeSizeTier(level int) error {
	merged, err := lsm.sstableManager.MergeSSTables(generateTableName(), lsm.tables[level])
	if err != nil {
		return err
	}
	lsm.tables[level+1] = append(lsm.tables[level+1], merged)
	// Remove the old tables from the current level
	for _, table := range lsm.tables[level] {
		table.Delete()
	}
	lsm.tables[level] = lsm.tables[level][:0] // Clear the current level tables
	return nil
}

func (lsm *LSM) shouldCompactLeveled(level int) bool {
	limit := float64(lsm.filePerLevel) * math.Pow(float64(lsm.multiplier), float64(level))
	return float64(len(lsm.tables[level])) > limit
}

func (lsm *LSM) mergeLeveled(level int) error {
	evicted := lsm.tables[level][0]
	// Remove the first table from the current level
	lsm.tables[level] = lsm.tables[level][1:]

	l, r := 0, -1
	i := len(lsm.tables[level+1]) - 1
	for ; i >= 0; i-- {
		overlap, err := evicted.AreKeysOverlaping(lsm.tables[level+1][i])
		if err != nil {
			return err
		}
		if overlap {
			r = i
			break
		}
	}
	if r == -1 {
		// No overlap found, we can just append the evicted table to the next level
		lsm.tables[level+1] = append(lsm.tables[level+1], evicted)
		return nil
	}
	for i = r - 1; i >= 0; i-- {
		overlap, err := evicted.AreKeysOverlaping(lsm.tables[level+1][i])
		if err != nil {
			return err
		}
		if !overlap {
			l = i + 1
			break
		}
	}
	if l == r {
		// If l and r are the same, we can just append the evicted table to the next level
		lsm.tables[level+1] = append(lsm.tables[level+1], evicted)
		return nil
	}
	readers := lsm.tables[level+1][l : r+1]
	merged, err := lsm.sstableManager.MergeSSTables(generateTableName(), readers)
	if err != nil {
		return err
	}
	// Remove the merged tables from the current level
	for i := l; i <= r; i++ {
		lsm.tables[level+1][i].Delete()
	}
	evicted.Delete()
	leftPart := lsm.tables[level+1][:l]
	rightPart := lsm.tables[level+1][r+1:]
	// Reconstruct the next level tables without the merged tables
	lsm.tables[level+1] = append(leftPart, merged)
	lsm.tables[level+1] = append(lsm.tables[level+1], rightPart...)
	return nil
}
