package externalconfiguration

type CompressionConfig struct {
	Enabled bool `json:"enabled"`
}

func (c *CompressionConfig) Validate(dbName string) {}
