package validation

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// CharacterEncoding checks for proper UTF-8 encoding (rule 2.3.18.1)
// Ensures no mojibake or encoding issues
func (r *Rules) CharacterEncoding(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.1",
		Name:   "Character encoding must be correct (UTF-8)",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title
	if hasEncodingIssues(actual.Title) {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title has character encoding issues: '%s'", actual.Title),
		})
	}

	// Check each track
	for _, track := range actual.Tracks {
		// Check track title
		if hasEncodingIssues(track.Title) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Title has character encoding issues: '%s'",
					formatTrackNumber(track), track.Title),
			})
		}

		// Check artist names
		for _, artist := range track.Artists {
			if hasEncodingIssues(artist.Name) {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelError,
					Track: track.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Artist name has character encoding issues: '%s'",
						formatTrackNumber(track), artist.Name),
				})
			}
		}

		// Check filename
		if hasEncodingIssues(track.Name) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Filename has character encoding issues: '%s'",
					formatTrackNumber(track), track.Name),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// hasEncodingIssues checks if a string has encoding problems
func hasEncodingIssues(s string) bool {
	if s == "" {
		return false
	}

	// Check if valid UTF-8
	if !utf8.ValidString(s) {
		return true
	}

	// Check for mojibake patterns (common encoding issues)
	// These are character sequences that indicate encoding problems
	mojibakePatterns := []string{
		"Ã©", "Ã¨", "Ã ", "Ã¤", "Ã¶", "Ã¼", "Ã±", // Latin-1 -> UTF-8 double encoding
		"â€™", "â€œ", "â€", // Smart quotes issues
		"Ã", "Â", // Common mojibake prefixes
	}

	for _, pattern := range mojibakePatterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}

	// Check for replacement character (indicates encoding failure)
	if strings.Contains(s, "\uFFFD") {
		return true
	}

	// Check for suspicious control characters (except newline, tab)
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return true
		}
	}

	// Check for NULL bytes
	if strings.Contains(s, "\x00") {
		return true
	}

	return false
}
