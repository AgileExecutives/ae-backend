package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove accents and diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove all non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

// EnsureUniqueSlug ensures a slug is unique by appending a number if necessary
func EnsureUniqueSlug(baseSlug string, existingSlugs []string) string {
	slug := baseSlug
	counter := 1

	// Create a map for faster lookup
	slugMap := make(map[string]bool)
	for _, s := range existingSlugs {
		slugMap[s] = true
	}

	// Keep incrementing until we find a unique slug
	for slugMap[slug] {
		slug = baseSlug + "-" + string(rune('0'+counter))
		counter++
	}

	return slug
}
