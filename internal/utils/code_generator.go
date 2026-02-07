package utils

import (
	"crypto/rand"
	"math/big"
)

// Character set for URL codes: uppercase letters and numbers only
const urlCodeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateURLCode generates a unique 11-character code for admin panel routing
// Format: [A-Z0-9]{11} (ex: FR34JJO390G)
// This code is NOT user-facing and provides security through obscurity for admin URLs
func GenerateURLCode() string {
	const codeLength = 11
	result := make([]byte, codeLength)

	charsetLen := big.NewInt(int64(len(urlCodeChars)))

	for i := 0; i < codeLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fallback to pseudo-random if crypto/rand fails
			// This should never happen in practice
			randomIndex = big.NewInt(int64(i % len(urlCodeChars)))
		}
		result[i] = urlCodeChars[randomIndex.Int64()]
	}

	return string(result)
}
