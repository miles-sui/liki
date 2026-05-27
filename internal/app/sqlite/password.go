package sqlite

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/25types/25types/internal/app/application/user"
	"golang.org/x/crypto/argon2"
)

// argon2id parameters.
const (
	argonTime    = 1
	argonMemory  = 47104
	argonThreads = 4
	argonSaltLen = 16
	argonKeyLen  = 32
	argonPrefix  = "$argon2id$v=19$m=47104,t=1,p=4$"
)

// PasswordHasher implements user.PasswordHasher using argon2id.
type PasswordHasher struct{}

var _ user.PasswordHasher = (*PasswordHasher)(nil)

// Hash creates an argon2id hash.
func (PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return argonPrefix + b64Salt + "$" + b64Hash, nil
}

// Verify checks password against a stored hash. Returns (valid, rehashIfParamsUpgraded).
func (PasswordHasher) Verify(password, storedHash string) (bool, string) {
	if storedHash == "" || !strings.HasPrefix(storedHash, "$argon2id$") {
		return false, ""
	}

	parts := strings.Split(storedHash, "$")
	if len(parts) < 6 {
		return false, ""
	}
	saltB64 := parts[len(parts)-2]
	hashB64 := parts[len(parts)-1]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, ""
	}
	expected, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, ""
	}

	computed := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	if subtle.ConstantTimeCompare(computed, expected) != 1 {
		return false, ""
	}

	// Signal rehash if parameters changed.
	if !strings.HasPrefix(storedHash, argonPrefix) {
		newHash, err := PasswordHasher{}.Hash(password)
		if err != nil {
			return true, ""
		}
		return true, newHash
	}
	return true, ""
}
