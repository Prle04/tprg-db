package wal

import (
	"db/components/blockmanager"
	"db/components/serialization"
	"time"
)

type walsegment struct {
	segmentNumber    uint64
	paddingThreshold uint64
	numberOfBlocks   uint64
	blockSize        uint64
	file             *blockmanager.BlockedFile
	size             uint64
}

func writeSegmentNumber(file *blockmanager.BlockedFile, segmentNumber uint64) error {
	err := file.Write(serialization.UInt64ToBytes(segmentNumber))
	if err != nil {
		return err
	}
	return file.Sync()
}

func newWALSegment(file *blockmanager.BlockedFile, numOfBlocks uint64, blockSize uint64, paddingThreshold uint64) (*walsegment, error) {
	file.Seek(0)
	SegmentNumberBytes, err := file.Read(uint64(SegmentNumberSize))
	if err != nil {
		return nil, err
	}
	size := uint64(SegmentNumberSize)
	segmentNumber := serialization.BytesToUInt64(SegmentNumberBytes)
	record := &walrecord{}
	// find the end of the segment
	for file.Tell() < numOfBlocks*blockSize {
		err := record.read(file)
		if err != nil {
			return nil, err
		}
		if record.recordType == FragmentTypePadding {
			file.Seek(file.Tell() - uint64(SegmentNumberSize+TypeSize))
			break
		}
		size += record.size()
	}
	return &walsegment{
		segmentNumber:    segmentNumber,
		file:             file,
		numberOfBlocks:   numOfBlocks,
		blockSize:        blockSize,
		size:             size,
		paddingThreshold: paddingThreshold,
	}, nil
}

// adds a record to the segment and returns leftover bytes or nil if the record was added successfully
func (segment *walsegment) add(key string, value []byte, isContinuance bool) []byte {
	maxSize := segment.numberOfBlocks * segment.blockSize
	record := &walrecord{
		recordType: FragmentTypeFull,
		key:        key,
		value:      make([]byte, 0),
		timestamp:  uint64(time.Now().Unix()),
		tombstone:  false,
	}
	// if the record is a continuation of a previous record we assume it is the last fragment
	if isContinuance {
		record.recordType = FragmentTypeLast
	}
	// check if the record fits in the segment
	if segment.size+record.size() > maxSize {
		return value
	}
	sizeLeft := maxSize - segment.size - record.size()
	inSegment := value[:min(sizeLeft, uint64(len(value)))]
	if len(value) != len(inSegment) && len(inSegment) < int(segment.paddingThreshold) {
		return value
	}
	record.value = append(record.value, inSegment...)
	segment.size += record.size()
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
	record.write(segment.file)
	return returnValue
}

// deletes a record from the segment and returns true if the record was deleted successfully
func (segment *walsegment) delete(key string) bool {
	record := &walrecord{
		key:       key,
		recordType: FragmentTypeFull,
		value:     make([]byte, 0),
		timestamp: uint64(time.Now().Unix()),
		tombstone: true,
	}
	// check if the record fits in the segment
	if segment.size+record.size() > segment.numberOfBlocks*segment.blockSize {
		return false
	}
	segment.size += record.size()
	record.write(segment.file)
	return true
}

func (segment *walsegment) close() {
	segment.file.Close()
}
