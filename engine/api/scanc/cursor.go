package scanc

import "tprgdb/engine/item"

// Remove direct import to avoid cycle
// import "tprgdb/engine"

type ICursor interface {
	GetPage(pageNumber int) ([]*item.DBItem, error)
	GetCurrentPage() int
	GetPageSize() int
	NextPage() ([]*item.DBItem, error)
	PreviousPage() ([]*item.DBItem, error)
}
