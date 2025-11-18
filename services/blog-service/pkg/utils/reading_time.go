package utils

import (
	"regexp"
	"strings"
)

const (
	// Average reading speed in words per minute
	wordsPerMinute = 200
)

// CalculateReadingTime calculates the estimated reading time in minutes
func CalculateReadingTime(content string) int {
	if content == "" {
		return 0
	}

	// Remove HTML tags
	content = stripHTMLTags(content)

	// Remove Markdown syntax
	content = stripMarkdownSyntax(content)

	// Count words
	wordCount := countWords(content)

	// Calculate reading time in minutes (minimum 1 minute)
	readingTime := wordCount / wordsPerMinute
	if readingTime < 1 {
		readingTime = 1
	}

	return readingTime
}

// stripHTMLTags removes HTML tags from the content
func stripHTMLTags(content string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(content, " ")
}

// stripMarkdownSyntax removes common Markdown syntax
func stripMarkdownSyntax(content string) string {
	// Remove code blocks
	re := regexp.MustCompile("```[\\s\\S]*?```")
	content = re.ReplaceAllString(content, " ")

	// Remove inline code
	re = regexp.MustCompile("`[^`]*`")
	content = re.ReplaceAllString(content, " ")

	// Remove links [text](url)
	re = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	content = re.ReplaceAllString(content, "$1")

	// Remove images ![alt](url)
	re = regexp.MustCompile(`!\[([^\]]*)\]\([^\)]+\)`)
	content = re.ReplaceAllString(content, "$1")

	// Remove headers (#, ##, ###, etc.)
	re = regexp.MustCompile(`#+\s*`)
	content = re.ReplaceAllString(content, "")

	// Remove bold/italic markers (**, *, __, _)
	re = regexp.MustCompile(`[*_]{1,2}([^*_]+)[*_]{1,2}`)
	content = re.ReplaceAllString(content, "$1")

	// Remove blockquotes
	re = regexp.MustCompile(`^>\s*`, )
	content = re.ReplaceAllString(content, "")

	// Remove list markers (-, *, +, 1., 2., etc.)
	re = regexp.MustCompile(`^[\s]*[-*+]\s+`)
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile(`^[\s]*\d+\.\s+`)
	content = re.ReplaceAllString(content, "")

	return content
}

// countWords counts the number of words in the content
func countWords(content string) int {
	// Split by whitespace and count non-empty strings
	words := strings.Fields(content)
	return len(words)
}
