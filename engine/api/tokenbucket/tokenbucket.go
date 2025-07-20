package tokenbucket

import (
	"errors"
	"time"
	"tprgdb/components/serialization"
	"tprgdb/engine/core"
)

const (
	TOKEN_BUCKET_PREFIX = "T"
)

type TokenBucket struct {
	core            *core.Core
	refreshDuration int64 // in seconds
	tokenCount      int64 // maximum tokens in the bucket
}

func NewTokenBucket(core *core.Core, refreshDuration int64, tokenCount int64) *TokenBucket {
	return &TokenBucket{
		core:            core,
		refreshDuration: refreshDuration,
		tokenCount:      tokenCount,
	}
}

func (tb *TokenBucket) Validate(username string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	token, err := tb.core.Get(TOKEN_BUCKET_PREFIX + username)
	if err != nil {
		return err
	}
	if token == nil || token.IsDeleted() {
		return errors.New("bucket not found for user: " + username)
	}
	bucket := &Bucket{}
	if err := bucket.DeserializeFromBytes(token.GetValue()); err != nil {
		return errors.New("failed to deserialize bucket for user: " + username + ", error: " + err.Error())
	}
	if time.Now().Unix()-bucket.LastRefresh > tb.refreshDuration {
		bucket.Tokens = tb.tokenCount // Reset tokens if the bucket hasn't been refreshed in the specified duration
		bucket.LastRefresh = time.Now().Unix()
	}
	if bucket.Tokens <= 0 {
		return errors.New("no tokens available for user: " + username)
	}
	bucket.Tokens-- // Consume a token
	tb.core.Set(TOKEN_BUCKET_PREFIX+username, bucket.SerializeToBytes())
	return nil
}

type Bucket struct {
	Tokens      int64
	LastRefresh int64
}

func (b *Bucket) SerializeToBytes() []byte {
	data := make([]byte, 0)
	data = append(data, serialization.Int64ToBytes(b.Tokens)...)
	data = append(data, serialization.Int64ToBytes(b.LastRefresh)...)
	return data
}

func (b *Bucket) DeserializeFromBytes(data []byte) error {
	if len(data) != 16 {
		return errors.New("data too short to deserialize Bucket")
	}
	b.Tokens = serialization.BytesToInt64(data[:8])
	b.LastRefresh = serialization.BytesToInt64(data[8:16])
	return nil
}
