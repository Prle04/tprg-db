package item

import "time"

type DBItem struct {
	Key       string
	Value     []byte
	Timestamp *time.Time
}

func (i *DBItem) GetKey() string {
	return i.Key
}
func (i *DBItem) GetValue() []byte {
	return i.Value
}
func (i *DBItem) GetTimestamp() *time.Time {
	return i.Timestamp
}
func (i *DBItem) GetUnixTimestamp() uint64 {
	return uint64(i.Timestamp.Unix())
}
func (i *DBItem) GetUnixNanoTimestamp() uint64 {
	return uint64(i.Timestamp.UnixNano())
}
func (i *DBItem) GetTimestampString() string {
	return i.Timestamp.Format("02.01.2006 15:04:05")
}
