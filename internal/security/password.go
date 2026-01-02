package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

const (
	// DefaultPasswordLength is the default length for generated passwords
	DefaultPasswordLength = 16

	// Charsets for password generation
	LowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	UppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DigitChars     = "0123456789"
	SpecialChars   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	AlphanumChars  = LowercaseChars + UppercaseChars + DigitChars
	AllChars       = AlphanumChars + SpecialChars
)

// GeneratePassword generates a cryptographically secure random password
func GeneratePassword(length int) (string, error) {
	if length <= 0 {
		length = DefaultPasswordLength
	}

	// Use base64 encoding of random bytes for simplicity
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 encode and take first 'length' characters
	encoded := base64.StdEncoding.EncodeToString(bytes)

	// Remove characters that might cause issues in URLs or configs
	encoded = strings.ReplaceAll(encoded, "+", "")
	encoded = strings.ReplaceAll(encoded, "/", "")
	encoded = strings.ReplaceAll(encoded, "=", "")

	if len(encoded) < length {
		// Recursively generate more if needed
		additional, err := GeneratePassword(length - len(encoded))
		if err != nil {
			return "", err
		}
		encoded += additional
	}

	return encoded[:length], nil
}

// GenerateAlphanumericPassword generates a password with only alphanumeric characters
func GenerateAlphanumericPassword(length int) (string, error) {
	return generateFromCharset(length, AlphanumChars)
}

// GenerateSecurePassword generates a password with mixed character types
func GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	// Ensure at least one of each type
	var password strings.Builder

	// Add at least one of each type
	lowercase, err := randomChar(LowercaseChars)
	if err != nil {
		return "", err
	}
	password.WriteString(lowercase)

	uppercase, err := randomChar(UppercaseChars)
	if err != nil {
		return "", err
	}
	password.WriteString(uppercase)

	digit, err := randomChar(DigitChars)
	if err != nil {
		return "", err
	}
	password.WriteString(digit)

	// Fill the rest with alphanumeric characters
	remaining := length - 3
	for i := 0; i < remaining; i++ {
		char, err := randomChar(AlphanumChars)
		if err != nil {
			return "", err
		}
		password.WriteString(char)
	}

	// Shuffle the password
	return shuffleString(password.String())
}

// generateFromCharset generates a random string from a given charset
func generateFromCharset(length int, charset string) (string, error) {
	var password strings.Builder
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password.WriteByte(charset[n.Int64()])
	}

	return password.String(), nil
}

// randomChar returns a random character from the given charset
func randomChar(charset string) (string, error) {
	return generateFromCharset(1, charset)
}

// shuffleString shuffles a string randomly
func shuffleString(s string) (string, error) {
	runes := []rune(s)
	n := len(runes)

	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("failed to shuffle: %w", err)
		}
		j := jBig.Int64()
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes), nil
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateBase64Token generates a base64-encoded random token
func GenerateBase64Token(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GenerateURLSafeToken generates a URL-safe base64-encoded random token
func GenerateURLSafeToken(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// IsStrongPassword checks if a password meets minimum strength requirements
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLower := false
	hasUpper := false
	hasDigit := false

	for _, c := range password {
		switch {
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}

	return hasLower && hasUpper && hasDigit
}
