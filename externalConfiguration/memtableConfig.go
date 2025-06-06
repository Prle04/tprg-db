package externalconfiguration

type MemtableConfig struct {
	DataStructure     string `json:"dataStructure"`
	NumberOfInstances int    `json:"numberOfInstances"`
	Threshld          int    `json:"threshld"`
}

const (
	SKIPLIST = "skiplist"
	HASHMAP  = "hashmap"
	BTREE    = "b-tree"
)

func (m *MemtableConfig) Validate() {
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
}
