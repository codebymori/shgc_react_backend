package utils

import (
	"regexp"
	"strings"
)

// GenerateSlug generates a URL-friendly slug from a title
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)
	
	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	
	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Replace multiple consecutive hyphens with single hyphen
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	
	return slug
}
