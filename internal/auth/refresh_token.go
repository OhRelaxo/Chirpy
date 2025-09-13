package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	key := make([]byte, 256)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	randStr := hex.EncodeToString(key)
	return randStr, nil
}
