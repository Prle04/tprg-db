package blockmanager

import (
	"encoding/binary"
	"os"
)

type BlockedFile struct {
	manager   *BlockManager
	path      string
	position  uint64
	blockSize uint64
}

func (f *BlockedFile) getBlockOrderNumber() uint64 {
	return f.position / f.manager.blockSize
}

func (f *BlockedFile) Tell() uint64 {
	return f.position
}

func (f *BlockedFile) Seek(offset uint64) {
	f.position = offset
}

func (f *BlockedFile) BlockSize() uint64 {
	return f.blockSize
}

func (f *BlockedFile) Read(length uint64) ([]byte, error) {
	orderNumber := f.getBlockOrderNumber()
	block, err := f.manager.getBlock(getBlockPath(f.path, orderNumber))
	if err != nil {
		return nil, err
	}
	data := make([]byte, 0)
	for length > 0 {
		remainingSize := block.getRemainingSize(f.position % f.manager.blockSize)
		readLength := min(length, remainingSize)
		readData, err := block.Read(f.position%f.manager.blockSize, readLength)
		if err != nil {
			return nil, err
		}
		data = append(data, readData...)
		length -= readLength
		f.position += readLength
		if readLength == 0 || length == 0 {
			break
		}
		orderNumber++
		block, err = f.manager.getBlock(getBlockPath(f.path, orderNumber))
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (f *BlockedFile) ReadUVariable() ([]byte, uint64, error) {
	data, err := f.Read(8) // Read 8 bytes for a uint64
	if err != nil {
		return nil, 0, err
	}
	value, n := binary.Uvarint(data)
	if n <= 0 {
		return nil, 0, &VariableLengthError{}
	}
	f.position -= 8 // Return to the start of the variable length data
	bytesRead, err := f.Read(uint64(n))
	if err != nil {
		return nil, 0, err
	}
	return bytesRead, value, nil
}

func (f *BlockedFile) ReadVariable() ([]byte, int64, error) {
	data, err := f.Read(8) // Read 8 bytes for a uint64
	if err != nil {
		return nil, 0, err
	}
	value, n := binary.Varint(data)
	if n <= 0 {
		return nil, 0, &VariableLengthError{}
	}
	f.position -= 8 // Return to the start of the variable length data
	bytesRead, err := f.Read(uint64(n))
	if err != nil {
		return nil, 0, err
	}
	return bytesRead, value, nil
}

func (f *BlockedFile) Write(data []byte) error {
	orderNumber := f.getBlockOrderNumber()
	block, err := f.manager.getBlock(getBlockPath(f.path, orderNumber))
	if err != nil {
		return err
	}
	for len(data) > 0 {
		remainingSize := block.getRemainingSize(f.position % f.manager.blockSize)
		length := min(uint64(len(data)), remainingSize)
		block.Write(f.position%f.manager.blockSize, data[:length])
		f.position += length
		// if all data is written, break
		if length == uint64(len(data)) {
			break
		}
		// if block is full, get next block
		data = data[length:]
		orderNumber++
		block, err = f.manager.getBlock(getBlockPath(f.path, orderNumber))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *BlockedFile) Close() error {
	return f.manager.closeFile(f.path)
}

func (f *BlockedFile) Sync() error {
	return f.manager.syncFile(f.path)
}

func (f *BlockedFile) Delete() error {
	return f.manager.deleteFile(f.path)
}

func (f *BlockedFile) Stat() (os.FileInfo, error) {
	err := f.Sync()
	if err != nil {
		return nil, err
	}
	return os.Stat(f.path)
}
