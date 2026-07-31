package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	legacySaltBytes = 16
	legacyKeyLength = 64
	legacyScryptN   = 16384
	legacyScryptR   = 8
	legacyScryptP   = 1
)

// HashPassword emits the exact salt:derived-key encoding used by the legacy
// Next.js credentials provider.
func HashPassword(password string) (string, error) {
	saltBytes := make([]byte, legacySaltBytes)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)
	derivedKey, err := derivePassword(password, salt)
	if err != nil {
		return "", fmt.Errorf("derive password key: %w", err)
	}
	return salt + ":" + hex.EncodeToString(derivedKey), nil
}

// VerifyPassword accepts only the legacy salt:derived-key encoding and uses
// constant-time comparison for the derived key.
func VerifyPassword(password, hashedPassword string) (bool, error) {
	salt, storedHex, ok := strings.Cut(hashedPassword, ":")
	if !ok || salt == "" || storedHex == "" {
		return false, nil
	}
	if len(salt) != legacySaltBytes*2 || !isLowerHex(salt) {
		return false, nil
	}
	storedKey, err := hex.DecodeString(storedHex)
	if err != nil || len(storedKey) != legacyKeyLength {
		return false, nil
	}
	derivedKey, err := derivePassword(password, salt)
	if err != nil {
		return false, fmt.Errorf("derive password key for verification: %w", err)
	}
	return subtle.ConstantTimeCompare(derivedKey, storedKey) == 1, nil
}

func derivePassword(password, salt string) ([]byte, error) {
	return scrypt.Key([]byte(password), []byte(salt), legacyScryptN, legacyScryptR, legacyScryptP, legacyKeyLength)
}

func isLowerHex(value string) bool {
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
