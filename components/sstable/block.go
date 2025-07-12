package sstable

import "tprgdb/components/blockmanager"

type sstableBlock struct {
	size uint64
	data []byte
}

func newSSTableBlock() *sstableBlock {
	return &sstableBlock{
		size: 0,
		data: make([]byte, 0),
	}
}

// adds a record to the segment and returns leftover bytes or nil if the record was added successfully
func (segment *sstableBlock) add(
	file *blockmanager.BlockedFile,
	key string,
	value []byte,
	timestamp uint64,
	isContinuance bool,
) ([]byte, error) {
	record := &sstablerecord{
		recordType: FragmentTypeFull,
		key:        key,
		value:      make([]byte, 0),
		timestamp:  timestamp,
		tombstone:  false,
	}
	// if the record is a continuation of a previous record we assume it is the last fragment
	if isContinuance {
		record.recordType = FragmentTypeLast
	}
	recordSize := record.size()
	// check if the record fits in the segment
	if segment.size+recordSize > SegmentBlockSize {
		return value, nil
	}
	sizeLeft := SegmentBlockSize - segment.size - recordSize
	inSegment := value[:min(sizeLeft, uint64(len(value)))]
	if len(value) != len(inSegment) && len(inSegment) < int(PaddingThreshold) {
		return value, nil
	}
	record.value = append(record.value, inSegment...)
	recordSize = record.size()
	segment.size += recordSize
	var returnValue []byte = nil
	if len(inSegment) < len(value) {
		returnValue = value[len(inSegment):]
		// we know that the record is not last fragment so we set the type to middle or first depending on the previous record
		if isContinuance {
			record.recordType = FragmentTypeMiddle
		} else {
			record.recordType = FragmentTypeFirst
		}

	}
	segment.data = append(segment.data, record.getData()...)
	record.write(file)
	return returnValue, nil
}

// deletes a record from the segment, returns true if the record was deleted successfully
func (segment *sstableBlock) delete(file *blockmanager.BlockedFile, key string, timestamp uint64) (bool, error) {
	record := &sstablerecord{
		key:        key,
		recordType: FragmentTypeFull,
		value:      make([]byte, 0),
		timestamp:  timestamp,
		tombstone:  true,
	}
	// check if the record fits in the segment
	if segment.size+record.size() > SegmentBlockSize {
		return false, nil
	}
	data := record.getData()
	segment.size += uint64(len(data))
	segment.data = append(segment.data, data...)
	err := record.write(file)
	return true, err
}

func (segment *sstableBlock) getData() []byte {
	return segment.data
}

func (segment *sstableBlock) clear() {
	segment.size = 0
	segment.data = make([]byte, 0)
}
