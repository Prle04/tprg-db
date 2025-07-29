package externalconfiguration

type CacheConfig struct {
	MaxElements int `json:"maxElements"`
}

func (c *CacheConfig) Validate(dbName string) {
	if c.MaxElements <= 0 {
		c.MaxElements = 1024 // Default value
	}
}
