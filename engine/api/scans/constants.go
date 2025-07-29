package scans

const (
	CACHE_SIZE           = 10 // Example cache size, adjust as needed
	ITERATOR_BUFFER_SIZE = 10 // Example buffer size for iterators
	PREFIX_SIZE          = 1  // Size of the prefix used for differentiating types of operations
)

type EndOfIterationError struct{}

func (e *EndOfIterationError) Error() string {
	return "end of iteration"
}

var END error = &EndOfIterationError{}
