package random

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	mRand "math/rand"
)

// GenerateString generates a random string of the specified length
func GenerateString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// GenerateSecretKey generates a 16-character secret key
func GenerateSecretKey() string {
	return GenerateString(16)
}

// GeneratePluginKey generates a 6-character plugin key
func GeneratePluginKey() string {
	return GenerateString(32)
}

func GenerateRandomString(length int) string {
	// Define the character set
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// Build the random string
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[mRand.Intn(len(charset))]
	}
	return string(result)
}

func GeneratePassword(length int) (string, error) {
	// Safe charset: uppercase and lowercase letters + digits + special characters
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789" +
		"!@#$%^&*()-_=+[]{}<>?/"

	pass := make([]byte, length)

	for i := 0; i < length; i++ {
		// Securely pick a random character from the charset
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		pass[i] = charset[n.Int64()]
	}

	return string(pass), nil
}
