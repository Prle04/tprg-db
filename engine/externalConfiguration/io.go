package externalconfiguration

import (
	"encoding/json"
	"os"
)

func ReadExternalConfiguration(file *os.File, dbName string) (*ExternalConfiguration, error) {
	var config ExternalConfiguration
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	config.Validate(dbName)
	return &config, nil
}

func WriteExternalConfiguration(file *os.File, config *ExternalConfiguration) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}
