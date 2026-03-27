package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

const (
	// base62Chars is the character set for auto-generated slugs.
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// slugLength is the number of characters in an auto-generated slug.
	slugLength = 6

	// minCustomSlugLen is the minimum length of a custom slug.
	minCustomSlugLen = 3

	// maxCustomSlugLen is the maximum length of a custom slug.
	maxCustomSlugLen = 64
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// GenerateSlug produces a cryptographically random 6-character Base62 slug.
func GenerateSlug() string {
	b := make([]byte, slugLength)
	max := big.NewInt(int64(len(base62Chars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// crypto/rand should never fail on a healthy system.
			panic(fmt.Sprintf("crypto/rand failed: %v", err))
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b)
}

// ValidateCustomSlug checks that a user-provided slug meets the business rules:
// 3–64 characters, alphanumeric and hyphens only.
func ValidateCustomSlug(slug string) error {
	if len(slug) < minCustomSlugLen {
		return fmt.Errorf("slug must be at least %d characters, got %d", minCustomSlugLen, len(slug))
	}
	if len(slug) > maxCustomSlugLen {
		return fmt.Errorf("slug must be at most %d characters, got %d", maxCustomSlugLen, len(slug))
	}
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("slug must contain only letters, numbers, and hyphens")
	}
	return nil
}
