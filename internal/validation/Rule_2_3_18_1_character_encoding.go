package validation

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// CharacterEncoding checks for proper UTF-8 encoding (album: 2.3.18.1-album, track: 2.3.18.1)
// Ensures no mojibake or encoding issues across album title/folder and track tags
func (r *Rules) AlbumCharacterEncoding(actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.1-album",
		Name:   "Character encoding must be correct (UTF-8)",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title
	if hasEncodingIssues(actualTorrent.Title) {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title has character encoding issues: '%s'", actualTorrent.Title),
		})
	}

	// Check album folder name
	if hasEncodingIssues(actualTorrent.RootPath) {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album folder name has character encoding issues: '%s'", actualTorrent.RootPath),
		})
	}

	return RuleResult{Meta: meta, Issues: issues}
}

func (r *Rules) TrackCharacterEncoding(actualTrack, _ *domain.Track, _, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.1",
		Name:   "Character encoding must be correct (UTF-8)",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue
	// Check track title
	if hasEncodingIssues(actualTrack.Title) {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Title has character encoding issues: '%s'",
				formatTrackNumber(actualTrack), actualTrack.Title),
		})
	}

	// Check artist names
	for _, artist := range actualTrack.Artists {
		if hasEncodingIssues(artist.Name) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Artist name has character encoding issues: '%s'",
					formatTrackNumber(actualTrack), artist.Name),
			})
		}
	}

	// Check filename
	if hasEncodingIssues(actualTrack.File.Path) {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Filename has character encoding issues: '%s'",
				formatTrackNumber(actualTrack), actualTrack.File.Path),
		})
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
