package externalconfiguration

type MemtableConfig struct {
	DataStructure     string `json:"dataStructure"`
	NumberOfInstances int    `json:"numberOfInstances"`
	Threshld          int    `json:"threshld"`

	SkipList SkipListConfig `json:"skipList"`
	BTree    BTreeConfig    `json:"bTree"`
}

const (
	SKIPLIST = "skiplist"
	HASHMAP  = "hashmap"
	BTREE    = "b-tree"
)

func (m *MemtableConfig) Validate(dbName string) {
	validDataStructures := map[string]bool{
		SKIPLIST: true,
		HASHMAP:  true,
		BTREE:    true,
	}
	if _, exists := validDataStructures[m.DataStructure]; !exists {
		m.DataStructure = HASHMAP // Default to HASHMAP if invalid
	}

	if m.NumberOfInstances <= 1 {
		m.NumberOfInstances = 2 // Default to 2 instance if invalid
	}
	if m.Threshld <= 0 {
		m.Threshld = 1024 // Default threshold
	}

	m.SkipList.Validate(dbName)
	m.BTree.Validate(dbName)

}

type SkipListConfig struct {
	MaxHeight int `json:"maxHeight"`
}

func (s *SkipListConfig) Validate(dbName string) {
	if s.MaxHeight <= 0 {
		s.MaxHeight = 16 // Default value
	}
}

type BTreeConfig struct {
	MaxChildren int `json:"maxChildren"`
}

func (s *BTreeConfig) Validate(dbName string) {
	if s.MaxChildren <= 0 {
		s.MaxChildren = 16 // Default value
	}
}
