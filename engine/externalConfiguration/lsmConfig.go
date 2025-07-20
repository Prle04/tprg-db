package externalconfiguration

import "path"

type LSMConfig struct {
	MaxLevels  int            `json:"maxLevels"`
	FolderPath string         `json:"folderPath"`
	Algorithm  string         `json:"algorithm"`
	SizeTier   SizeTierConfig `json:"sizeTierConfig"`
	Leved      LevedConfig    `json:"leveledConfig"`
}

type SizeTierConfig struct {
	FilesPerLevel int `json:"filesPerLevel"`
}

type LevedConfig struct {
	InitialLevelSize int `json:"initialLevelSize"`
	LevelMultiplier  int `json:"levelMultiplier"`
}

func (l *LSMConfig) Validate(dbName string) {
	if l.MaxLevels <= 0 {
		l.MaxLevels = 5 // Default value
	}

	if l.FolderPath == "" {
		l.FolderPath = path.Join("dbs", dbName, "data", "lsm") // Default folder path
	}

	validAlgorithms := map[string]bool{
		"size-tier": true,
		"leveled":   true,
	}
	if _, exists := validAlgorithms[l.Algorithm]; !exists {
		l.Algorithm = "size-tier" // Default to size_tiered if invalid
	}
}

func (s *SizeTierConfig) Validate(dbName string) {
	if s.FilesPerLevel <= 0 {
		s.FilesPerLevel = 10 // Default value
	}
}

func (l *LevedConfig) Validate(dbName string) {
	if l.InitialLevelSize <= 0 {
		l.InitialLevelSize = 5 // Default value
	}
	if l.LevelMultiplier <= 1 {
		l.LevelMultiplier = 10 // Default multiplier
	}
}
