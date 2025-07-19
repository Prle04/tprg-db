package scanc

import "tprgdb/engine"

type ICursor interface {
	GetPage(pageNumber int) ([]*engine.FinalItem, error)
	GetCurrentPage() int
	GetPageSize() int
	NextPage() ([]*engine.FinalItem, error)
	PreviousPage() ([]*engine.FinalItem, error)
}
