package utils

import (
	"html"
	"strings"
)

// SanitizeHTML removes potentially dangerous HTML content
func SanitizeHTML(content string) string {
	// Escape all HTML entities to prevent XSS
	content = html.EscapeString(content)

	// Limit length
	if len(content) > 1000 {
		content = content[:1000]
	}

	return strings.TrimSpace(content)
}

// SanitizeTitle removes potentially dangerous content from titles
func SanitizeTitle(title string) string {
	// Escape HTML
	title = html.EscapeString(title)

	// Limit length
	if len(title) > 100 {
		title = title[:100]
	}

	return strings.TrimSpace(title)
}
