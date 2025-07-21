package scans

import (
	"tprgdb/engine/core"
	"tprgdb/engine/item"
)

type PrefixScanCursor struct {
	rangeScanCursor *RangeScanCursor
	prefix          string
}

func NewPrefixScanCursor(core *core.Core, prefix string, pageSize int) *PrefixScanCursor {
	return &PrefixScanCursor{
		rangeScanCursor: NewRangeScanCursor(core, prefix, prefix+"\uffff", pageSize),
		prefix:          prefix,
	}
}

func (c *PrefixScanCursor) GetPage(pageNumber int) ([]*item.DBItem, error) {
	return c.rangeScanCursor.GetPage(pageNumber)
}

func (c *PrefixScanCursor) GetPageSize() int {
	return c.rangeScanCursor.GetPageSize()
}

func (c *PrefixScanCursor) GetCurrentPage() int {
	return c.rangeScanCursor.GetCurrentPage()
}
func (c *PrefixScanCursor) NextPage() ([]*item.DBItem, error) {
	return c.rangeScanCursor.NextPage()
}
func (c *PrefixScanCursor) PreviousPage() ([]*item.DBItem, error) {
	return c.rangeScanCursor.PreviousPage()
}
