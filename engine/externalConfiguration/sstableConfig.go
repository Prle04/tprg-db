package externalconfiguration

type SSTableConfig struct {
	FolderPath            string `json:"folderPath"`
	SummaryDilutionFactor int    `json:"summaryDilutionFactor"`
	Mode                  string `json:"mode"`
}

func (s *SSTableConfig) Validate() {
	if s.FolderPath == "" {
		s.FolderPath = "data/sstable" // Default folder path
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
