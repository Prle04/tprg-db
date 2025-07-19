package externalconfiguration

type WALConfig struct {
	NumberOfBlocksPerSegment int    `json:"numberOfBlocksPerSegment"`
	WALFolder                string `json:"walFolder"`
}

func (w *WALConfig) Validate() {
	if w.NumberOfBlocksPerSegment <= 0 {
		w.NumberOfBlocksPerSegment = 2048 // Default value
	}
	if w.WALFolder == "" {
		w.WALFolder = "data/wal" // Default folder
	}
}
