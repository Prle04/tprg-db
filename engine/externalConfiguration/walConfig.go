package externalconfiguration

import "path"

type WALConfig struct {
	NumberOfBlocksPerSegment int    `json:"numberOfBlocksPerSegment"`
	WALFolder                string `json:"walFolder"`
}

func (w *WALConfig) Validate(dbName string) {
	if w.NumberOfBlocksPerSegment <= 0 {
		w.NumberOfBlocksPerSegment = 2048 // Default value
	}
	if w.WALFolder == "" {
		w.WALFolder = path.Join("dbs", dbName, "data", "wal") // Default folder
	}
}
