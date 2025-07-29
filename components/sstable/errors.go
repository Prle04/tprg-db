package sstable

type CRCError struct{}

func (e *CRCError) Error() string { return "CRC check failed" }

type NotFirstFragment struct {
	fragmentType uint8
}

func (e *NotFirstFragment) Error() string {
	typeName := ""
	switch e.fragmentType {
	case FragmentTypeMiddle:
		typeName = "Middle"
	case FragmentTypeLast:
		typeName = "Last"
	}
	return "First fragment of first segment cannot be of type " + typeName
}

type IncorrectModeError struct {
	mode string
}

func (e *IncorrectModeError) Error() string {
	return "Incorrect mode: " + e.mode + ", expected 'one-file' or 'multi-file'"
}

type EmptyMerkleTreeError struct{}

func (e *EmptyMerkleTreeError) Error() string {
	return "Merkle tree cannot be empty"
}

type EOFError struct{}

func (e *EOFError) Error() string {
	return "End of file reached"
}
