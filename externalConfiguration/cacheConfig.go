package externalconfiguration

type CacheConfig struct {
	MaxElements int `json:"maxElements"`
}

func (c *CacheConfig) Validate() {
	if c.MaxElements <= 0 {
		c.MaxElements = 1024 // Default value
	}
}
