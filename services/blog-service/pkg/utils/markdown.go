package utils

import (
	"bytes"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// MarkdownToHTML converts Markdown content to HTML
func MarkdownToHTML(markdown string) string {
	if markdown == "" {
		return ""
	}

	// Configure Blackfriday renderer with safe HTML options
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags |
			blackfriday.HTMLFlagsNone,
	})

	// Configure Markdown parser extensions
	extensions := blackfriday.CommonExtensions |
		blackfriday.AutoHeadingIDs |
		blackfriday.Footnotes |
		blackfriday.NoEmptyLineBeforeBlock

	// Parse and render
	html := blackfriday.Run([]byte(markdown), blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(extensions))

	return string(html)
}

// GenerateExcerpt generates an excerpt from HTML content
func GenerateExcerpt(html string, maxLength int) string {
	if html == "" {
		return ""
	}

	// Strip HTML tags
	text := stripHTMLTags(html)

	// Trim whitespace
	text = strings.TrimSpace(text)

	// Truncate to max length
	if len(text) <= maxLength {
		return text
	}

	// Find the last space before maxLength to avoid cutting words
	excerpt := text[:maxLength]
	lastSpace := strings.LastIndex(excerpt, " ")
	if lastSpace > 0 {
		excerpt = excerpt[:lastSpace]
	}

	return excerpt + "..."
}

// SanitizeMarkdown removes potentially dangerous content from Markdown
func SanitizeMarkdown(markdown string) string {
	// Remove script tags
	markdown = removeScriptTags(markdown)

	// Remove iframe tags
	markdown = removeIframeTags(markdown)

	return markdown
}

func removeScriptTags(content string) string {
	var result bytes.Buffer
	inScript := false

	for i := 0; i < len(content); i++ {
		if i+8 < len(content) && strings.ToLower(content[i:i+8]) == "<script" {
			inScript = true
		}

		if !inScript {
			result.WriteByte(content[i])
		}

		if i+9 < len(content) && strings.ToLower(content[i:i+9]) == "</script>" {
			inScript = false
			i += 8 // Skip the closing tag
		}
	}

	return result.String()
}

func removeIframeTags(content string) string {
	var result bytes.Buffer
	inIframe := false

	for i := 0; i < len(content); i++ {
		if i+7 < len(content) && strings.ToLower(content[i:i+7]) == "<iframe" {
			inIframe = true
		}

		if !inIframe {
			result.WriteByte(content[i])
		}

		if i+9 < len(content) && strings.ToLower(content[i:i+9]) == "</iframe>" {
			inIframe = false
			i += 8 // Skip the closing tag
		}
	}

	return result.String()
}
