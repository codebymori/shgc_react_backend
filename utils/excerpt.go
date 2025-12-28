package utils

import (
	"regexp"
	"strings"
)

// MakeExcerpt generates a plain text excerpt from HTML content
// It strips HTML tags and limits the text to the specified character limit
func MakeExcerpt(html string, limit int) string {
	if html == "" {
		return ""
	}

	// Replace block-level tags with spaces to preserve word boundaries
	blockTags := regexp.MustCompile(`<(\/)?(p|br|div|h[1-6]|li|ol|ul)[^>]*>`)
	withSpaces := blockTags.ReplaceAllString(html, " ")

	// Strip all remaining HTML tags
	allTags := regexp.MustCompile(`<[^>]+>`)
	text := allTags.ReplaceAllString(withSpaces, "")

	// Clean up multiple whitespaces and trim
	multiSpace := regexp.MustCompile(`\s+`)
	clean := multiSpace.ReplaceAllString(text, " ")
	clean = strings.TrimSpace(clean)

	// Truncate to limit if necessary
	if len(clean) > limit {
		return clean[:limit] + "..."
	}

	return clean
}
