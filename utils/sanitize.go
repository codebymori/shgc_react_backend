package utils

import (
	"github.com/microcosm-cc/bluemonday"
)

// SanitizeHTML sanitizes HTML content to prevent XSS attacks
// while allowing safe formatting tags commonly used by rich text editors like React Quill
func SanitizeHTML(html string) string {
	// Use UGC (User Generated Content) policy which allows common formatting
	// but blocks dangerous elements and attributes
	policy := bluemonday.UGCPolicy()
	
	// The UGC policy already allows most React Quill tags:
	// - Headings: h1-h6
	// - Paragraphs: p
	// - Text formatting: strong, em, u, s, del, ins, sub, sup
	// - Lists: ul, ol, li
	// - Links: a (with href)
	// - Images: img (with src, alt)
	// - Quotes: blockquote
	// - Code: code, pre
	// - Line breaks: br
	// - Tables: table, thead, tbody, tr, th, td
	
	// Sanitize and return
	return policy.Sanitize(html)
}
