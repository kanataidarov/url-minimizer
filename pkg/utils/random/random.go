package random

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomString(length int64) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
