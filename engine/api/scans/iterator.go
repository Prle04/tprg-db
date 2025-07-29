package scans

import "tprgdb/engine/item"

type IIterator interface {
	Next() (*item.DBItem, error)
	Stop()
}
