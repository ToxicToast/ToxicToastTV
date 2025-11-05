package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Remove accents and diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	slug, _, _ = transform.String(t, slug)

	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Remove consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	// Truncate to 200 characters
	if len(slug) > 200 {
		slug = slug[:200]
		// Remove trailing hyphen if any
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// GenerateUniqueSlug generates a unique slug by appending a number if necessary
func GenerateUniqueSlug(base string, exists func(string) bool) string {
	slug := GenerateSlug(base)

	if !exists(slug) {
		return slug
	}

	// Try appending numbers until we find a unique slug
	for i := 2; i < 1000; i++ {
		candidateSlug := fmt.Sprintf("%s-%d", slug, i)
		if !exists(candidateSlug) {
			return candidateSlug
		}
	}

	// Fallback: use timestamp
	return fmt.Sprintf("%s-%d", slug, nowUnix())
}

// Helper to get current Unix timestamp (can be mocked in tests)
var nowUnix = func() int64 {
	return time.Now().Unix()
}
