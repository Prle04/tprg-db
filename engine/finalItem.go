package engine

import "time"

type FinalItem struct {
	Key       string
	Value     []byte
	Timestamp *time.Time
}
func (i *FinalItem) GetKey() string {
	return i.Key
}
func (i *FinalItem) GetValue() []byte {
	return i.Value
}
func (i *FinalItem) GetTimestamp() *time.Time {
	return i.Timestamp
}
func (i *FinalItem) GetUnixTimestamp() uint64 {
	return uint64(i.Timestamp.Unix())
}
func (i *FinalItem) GetUnixNanoTimestamp() uint64 {
	return uint64(i.Timestamp.UnixNano())
}
func (i *FinalItem) GetTimestampString() string {
	return i.Timestamp.Format("02.01.2006")
}