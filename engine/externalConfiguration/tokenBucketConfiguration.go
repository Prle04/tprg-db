package externalconfiguration

type TokenBucketConfiguration struct {
	RefreshDuration int64 `json:"refreshDuration"` // in seconds
	TokenCount      int64 `json:"tokenCount"`      // maximum tokens in the bucket
}

func (tbc *TokenBucketConfiguration) Validate(dbName string) {
	if tbc.RefreshDuration <= 0 {
		tbc.RefreshDuration = 60 * 15 // Default to 15 minutes if not set
	}
	if tbc.TokenCount <= 0 {
		tbc.TokenCount = 100 // Default to 100 tokens if not set
	}
}
