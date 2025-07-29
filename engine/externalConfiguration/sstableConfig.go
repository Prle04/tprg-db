package externalconfiguration

import "path"

type SSTableConfig struct {
	FolderPath            string `json:"folderPath"`
	SummaryDilutionFactor int    `json:"summaryDilutionFactor"`
	Mode                  string `json:"mode"`
}

func (s *SSTableConfig) Validate(dbName string) {
	if s.FolderPath == "" {
		s.FolderPath = path.Join("dbs", dbName, "data", "sstables") // Default folder path
	}

	if s.SummaryDilutionFactor <= 0 {
		s.SummaryDilutionFactor = 10 // Default value
	}

	validModes := map[string]bool{
		"one-file":   true,
		"multi-file": true,
	}
	if _, exists := validModes[s.Mode]; !exists {
		s.Mode = "one-file" // Default to read mode if invalid
	}
}
