package scanc

const (
	CACHE_SIZE           = 10 // Example cache size, adjust as needed
	ITERATOR_BUFFER_SIZE = 10 // Example buffer size for iterators
)

type EndOfIterationError struct{}

func (e *EndOfIterationError) Error() string {
	return "end of iteration"
}

var END error = &EndOfIterationError{}
