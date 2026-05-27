package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken returns a cryptographically random hex-encoded token.
func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
