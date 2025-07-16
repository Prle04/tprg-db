package lsm

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateTableName() string {
	timestamp := time.Now().Format("20060102_150405")
	randomPart := generateRandomString(5)
	return "table_" + timestamp + "_" + randomPart
}
