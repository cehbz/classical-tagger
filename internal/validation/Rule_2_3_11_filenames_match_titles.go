package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// filenameTrackPattern extracts the title portion from filename
// Matches: "01 - Title.flac" or "01. Title.flac" or "01_Title.flac"
var filenameTrackPattern = regexp.MustCompile(`^\d+[\s\-_\.]+(.+?)\.[\w]+$`)

// FilenamesMatchTitles checks that filenames accurately reflect track titles (rule 2.3.11)
func (r *Rules) FilenamesMatchTitles(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.11",
		Name:   "Filenames must accurately reflect track titles",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	for _, track := range actual.Tracks {
		fileName := track.Name
		if fileName == "" {
			continue // Will be caught by other rules
		}

		trackTitle := track.Title
		if trackTitle == "" {
			continue // Will be caught by RequiredTags rule
		}

		// Extract just the filename (not path)
		parts := strings.Split(fileName, "/")
		justFileName := parts[len(parts)-1]

		// Extract the title portion from filename
		matches := filenameTrackPattern.FindStringSubmatch(justFileName)
		if len(matches) < 2 {
			// Can't parse filename - might be "track.flac" format
			continue
		}

		fileTitle := matches[1]

		// Normalize both titles for comparison
		normalizedFileTitle := normalizeTitle(fileTitle)
		normalizedTrackTitle := normalizeTitle(trackTitle)

		// Check for similarity
		if !titlesMatch(normalizedFileTitle, normalizedTrackTitle) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: track.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Filename '%s' does not match track title '%s'",
					formatTrackNumber(track), fileTitle, trackTitle),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// normalizeTitle normalizes a title for comparison
func normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)

	// Remove common punctuation that might differ
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, ",", "")
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, "'", "")
	normalized = strings.ReplaceAll(normalized, "\"", "")
	normalized = strings.ReplaceAll(normalized, "!", "")
	normalized = strings.ReplaceAll(normalized, "?", "")
	normalized = strings.ReplaceAll(normalized, "(", "")
	normalized = strings.ReplaceAll(normalized, ")", "")
	normalized = strings.ReplaceAll(normalized, "[", "")
	normalized = strings.ReplaceAll(normalized, "]", "")

	// Replace multiple spaces with single space
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// titlesMatch checks if two normalized titles match closely enough
func titlesMatch(title1, title2 string) bool {
	// Exact match
	if title1 == title2 {
		return true
	}

	// Check if one is a substring of the other (handles abbreviations)
	// "Symphony No 5" matches "Symphony No 5 in C Minor"
	if strings.Contains(title1, title2) || strings.Contains(title2, title1) {
		return true
	}

	// Check edit distance for minor differences (typos, slight misspellings)
	if levenshteinDistance(title1, title2) <= 3 {
		return true
	}

	return false
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
