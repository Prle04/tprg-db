package scanc

import "tprgdb/engine"

type IIterator interface {
	Next() (*engine.FinalItem, error)
	Stop()
}
