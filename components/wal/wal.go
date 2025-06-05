package wal

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"tprgdb/components/blockmanager"
)

type WriteAheadLog struct {
	segmentSize      uint64
	currentSegment   *walsegment
	walPath          string
	blockSize        uint64
	paddingThreshold uint64
	manager          *blockmanager.BlockManager
}

func NewWriteAheadLog(
	walPath string, segmentSize uint64, blockSize uint64,
	paddingThreshold uint64, manager *blockmanager.BlockManager) (*WriteAheadLog, error) {

	// read all the files in the directory
	entries, err := os.ReadDir(walPath)
	if err != nil {
		return nil, err
	}
	// filter out the files that are not wal files
	// FILE FORMAT: wal_%0nd.wal
	filepaths := make([]string, 0)
	for _, v := range entries {
		// skip directories
		if v.IsDir() {
			continue
		}
		// skip files that are not wal files in the format wal_%0nd.wal
		if filepath.Ext(v.Name()) != ".wal" && !strings.HasPrefix(v.Name(), "wal_") {
			continue
		}
		filepaths = append(filepaths, filepath.Join(walPath, v.Name()))
	}

	// sort the filepaths in ascending order
	sort.Slice(filepaths, func(i, j int) bool {
		return filepaths[i] < filepaths[j]
	})

	// create the wal object
	wal := &WriteAheadLog{
		segmentSize:      segmentSize,
		walPath:          walPath,
		blockSize:        blockSize,
		paddingThreshold: paddingThreshold,
		manager:          manager,
	}

	if len(filepaths) != 0 {
		// open the last segment
		segmentFile := manager.OpenBlockedFile(filepaths[len(filepaths)-1])
		currentSegment, segErr := newWALSegment(segmentFile, segmentSize, blockSize, paddingThreshold)
		if segErr != nil {
			return nil, segErr
		}
		wal.currentSegment = currentSegment
	} else {
		// there are no segments so we create a new one
		err := wal.openNewSegment(1)
		if err != nil {
			return nil, err
		}
	}

	return wal, nil
}

func (wal *WriteAheadLog) Add(key string, value []byte) error {
	isContinuance := false
	for len(value) > 0 {
		// add the data to the current segment and get the remaining data for which no space was available
		remaining := wal.currentSegment.add(key, value, isContinuance)
		if remaining == nil {
			break
		}
		// we need to open a new segment
		segmentNumber := wal.currentSegment.segmentNumber
		wal.currentSegment.close()
		err := wal.openNewSegment(segmentNumber + 1)
		if err != nil {
			return err
		}
		isContinuance = true
		value = remaining
	}
	return nil
}

func (wal *WriteAheadLog) Delete(key string) error {
	// add new DELETE record to the current segment
	isDeleted := wal.currentSegment.delete(key)
	if isDeleted {
		return nil
	}
	// if the record does not fit in the current segment, we need to open a new one
	segmentNumber := wal.currentSegment.segmentNumber
	wal.currentSegment.file.Close()
	err := wal.openNewSegment(segmentNumber + 1)
	if err != nil {
		return err
	}
	// add the DELETE record to the new segment
	wal.currentSegment.delete(key)
	return nil
}

func (wal *WriteAheadLog) openNewSegment(segmentNumber uint64) error {
	format := fmt.Sprintf("%%0%dd", WALLeadingZeroes)
	numberWithLeading := fmt.Sprintf(format, segmentNumber)
	segmentFile := wal.manager.OpenBlockedFile(path.Join(wal.walPath, "wal_"+numberWithLeading+".wal"))
	err := writeSegmentNumber(segmentFile, segmentNumber)
	if err != nil {
		return err
	}
	currentSegment, segErr := newWALSegment(segmentFile, wal.segmentSize, wal.blockSize, wal.paddingThreshold)
	if segErr != nil {
		return segErr
	}
	wal.currentSegment = currentSegment
	return nil
}

func (wal *WriteAheadLog) Close() error {
	return wal.currentSegment.file.Close()
}

func (wal *WriteAheadLog) GetCurrentSegmentNumber() uint64 {
	return wal.currentSegment.segmentNumber
}

type WALIterator struct {
	Record *walrecord
	Err    error
}

func (wal *WriteAheadLog) Iterate() <-chan *WALIterator {
	ch := make(chan *WALIterator)
	go func() {
		defer close(ch)
		wal.currentSegment.file.Sync()
		entries, err := os.ReadDir(wal.walPath)
		if err != nil {
			ch <- &WALIterator{Err: err}
			return
		}
		var lastRecord *walrecord = nil
		for _, v := range entries {
			if v.IsDir() {
				continue
			}
			if filepath.Ext(v.Name()) != ".wal" && !strings.HasPrefix(v.Name(), "wal_") {
				continue
			}
			filepath := path.Join(wal.walPath, v.Name())
			file := wal.manager.OpenBlockedFile(filepath)
			file.Seek(uint64(SegmentNumberSize))
			for file.Tell() < wal.segmentSize*wal.blockSize {
				record := &walrecord{}
				err := record.read(file)
				if err != nil {
					ch <- &WALIterator{Err: err}
					return
				}
				if record.recordType == FragmentTypePadding {
					break
				}
				if record.recordType == FragmentTypeMiddle || record.recordType == FragmentTypeLast {
					if lastRecord == nil {
						ch <- &WALIterator{Err: &NotFirstFragment{fragmentType: record.recordType}}
						return
					}
					lastRecord.value = append(lastRecord.value, record.value...)
					if record.recordType == FragmentTypeLast {
						ch <- &WALIterator{Record: lastRecord}
						lastRecord = nil
					}
				} else {
					lastRecord = record
				}
			}
		}
	}()
	return ch
}
