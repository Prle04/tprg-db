package memtable

type MemtableFlushRequiredError struct {
}

func (e *MemtableFlushRequiredError) Error() string {
	return "memtable flush required"
}