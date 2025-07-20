package serialization

import (
	"encoding/binary"
	"math"
)

func UInt8ToBytes(value uint8) []byte {
	return []byte{value}
}

func Int8ToBytes(value int8) []byte {
	return []byte{byte(value)}
}

func UInt16ToBytes(value uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, value)
	return bytes
}

func Int16ToBytes(value int16) []byte {
	return UInt16ToBytes(uint16(value))
}

func UInt32ToBytes(value uint32) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	return bytes
}

func Int32ToBytes(value int32) []byte {
	return UInt32ToBytes(uint32(value))
}

func UInt64ToBytes(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)
	return bytes
}

func Int64ToBytes(value int64) []byte {
	return UInt64ToBytes(uint64(value))
}

func Float32ToBytes(value float32) []byte {
	return UInt32ToBytes(math.Float32bits(value))
}

func Float64ToBytes(value float64) []byte {
	return UInt64ToBytes(math.Float64bits(value))
}

func StringToBytes(value string) []byte {
	return []byte(value)
}

func BoolToBytes(value bool) []byte {
	if value {
		return UInt8ToBytes(1)
	}
	return UInt8ToBytes(0)
}

func BytesToBool(bytes []byte) bool {
	intValue := BytesToUInt8(bytes)
	return intValue == 1
}

func BytesToUInt8(bytes []byte) uint8 {
	return uint8(bytes[0])
}

func BytesToInt8(bytes []byte) int8 {
	return int8(bytes[0])
}

func BytesToUInt16(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

func BytesToInt16(bytes []byte) int16 {
	return int16(BytesToUInt16(bytes))
}

func BytesToUInt32(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(bytes)
}

func BytesToInt32(bytes []byte) int32 {
	return int32(BytesToUInt32(bytes))
}

func BytesToUInt64(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

func BytesToInt64(bytes []byte) int64 {
	return int64(BytesToUInt64(bytes))
}

func BytesToFloat32(bytes []byte) float32 {
	return math.Float32frombits(BytesToUInt32(bytes))
}

func BytesToFloat64(bytes []byte) float64 {
	return math.Float64frombits(BytesToUInt64(bytes))
}

func BytesToString(bytes []byte) string {
	return string(bytes)
}

func UVarInt(value uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, value)
	return buf[:n]
}

func VarInt(value int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, value)
	return buf[:n]
}
