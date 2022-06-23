package gocloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

// HashBody returns the hex-encoded SHA-256 sum of ' body'
func HashBody(ctx context.Context, body []byte) (string, error) {

	hash := sha256.Sum256(body)
	str_hash := hex.EncodeToString(hash[:])

	return str_hash, nil
}
