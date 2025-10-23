package scraping

import (
	"html"
	"regexp"
	"strings"
)

// decodeHTMLEntities decodes HTML entities to their Unicode equivalents.
// Uses Go's standard html package for proper decoding.
func decodeHTMLEntities(s string) string {
	// First use html.UnescapeString for standard entities
	s = html.UnescapeString(s)
	
	// Handle any remaining problematic UTF-8 encoding issues
	// (e.g., "Ã«" -> "ë" - malformed UTF-8 that appears as double-encoded)
	replacements := map[string]string{
		"Ã«": "ë",
		"Ã¶": "ö",
		"Ã¼": "ü",
		"Ã¤": "ä",
		"Ã©": "é",
		"Ã¨": "è",
		"Ãª": "ê",
		"Ã®": "î",
		"Ã´": "ô",
		"Ã»": "û",
		"Ã§": "ç",
		"Ã±": "ñ",
		"Ã ": "à",
		"Ãœ": "Ü",
		"Ã‰": "É",
		"Ã€": "À",
		// Common malformed sequences
		"NoÃ«l": "Noël",
	}
	
	for malformed, correct := range replacements {
		s = strings.ReplaceAll(s, malformed, correct)
	}
	
	return s
}

// stripHTMLTags removes all HTML tags from a string.
func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// cleanWhitespace cleans up excessive whitespace in strings.
func cleanWhitespace(s string) string {
	// Replace multiple spaces with single space
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// toTitleCase converts ALL CAPS to Title Case while preserving some exceptions.
func toTitleCase(s string) string {
	words := strings.Fields(s)
	
	// Articles and prepositions that should stay lowercase (unless at start)
	lowercase := map[string]bool{
		"a": true, "an": true, "the": true,
		"and": true, "or": true, "but": true,
		"of": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "with": true,
		"de": true, "la": true, "le": true, "von": true, "van": true,
	}
	
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		
		wordLower := strings.ToLower(word)
		
		// First word always capitalized
		if i == 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			continue
		}
		
		// Check if it should be lowercase
		if lowercase[wordLower] && len(word) <= 3 {
			words[i] = wordLower
		} else {
			// Title case: First letter upper, rest lower
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	
	return strings.Join(words, " ")
}

// normalizeWhitespace normalizes various whitespace characters to standard space.
func normalizeWhitespace(s string) string {
	// Replace various whitespace characters with standard space
	replacements := map[rune]rune{
		'\t':     ' ', // tab
		'\n':     ' ', // newline
		'\r':     ' ', // carriage return
		'\u00A0': ' ', // non-breaking space
		'\u2009': ' ', // thin space
		'\u200B': ' ', // zero-width space
	}
	
	result := []rune{}
	for _, r := range s {
		if replacement, ok := replacements[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}
	
	// Clean up multiple spaces
	return cleanWhitespace(string(result))
}

// sanitizeText performs a complete text sanitization pipeline.
// Useful for cleaning up text extracted from HTML.
func sanitizeText(s string) string {
	// Remove HTML tags
	s = stripHTMLTags(s)
	
	// Decode HTML entities
	s = decodeHTMLEntities(s)
	
	// Normalize whitespace
	s = normalizeWhitespace(s)
	
	return s
}

// cleanHTMLEntities is a legacy alias for decodeHTMLEntities.
// Deprecated: Use decodeHTMLEntities instead.
func cleanHTMLEntities(s string) string {
	return decodeHTMLEntities(s)
}