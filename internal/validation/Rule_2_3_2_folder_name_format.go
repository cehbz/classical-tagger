package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// folderNamePattern matches the standard folder format
// Format: "Artist - Album [Year] [Format] [Extra Info]"
// Example: "Beethoven - Symphony No. 5 [1963] [FLAC]"
var folderNamePattern = regexp.MustCompile(`^.+\s+-\s+.+\s+\[\d{4}\]`)

// FolderNameFormat checks that album title follows proper folder naming format (rule 2.3.2)
// Format: "Artist - Album [Year] [Format] [Extra Info]"
func (r *Rules) FolderNameFormat(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.2",
		Name:   "Folder name format: Artist - Album [Year] [Format]",
		Level:  domain.LevelWarning, // Warning because format varies
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	albumTitle := actual.Title
	if albumTitle == "" {
		// Will be caught by RequiredTags
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check for basic structure: should have " - " separator
	if !strings.Contains(albumTitle, " - ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' should contain ' - ' separator between artist and title", albumTitle),
		})
		return RuleResult{Meta: meta, Issues: issues}
	}

	// Check for year in brackets
	yearPattern := regexp.MustCompile(`\[(\d{4})\]`)
	yearMatches := yearPattern.FindStringSubmatch(albumTitle)

	if len(yearMatches) == 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' should include year in brackets [YYYY]", albumTitle),
		})
	} else {
		// Verify year matches album year
		yearInTitle, _ := strconv.Atoi(yearMatches[1])
		albumYear := actual.OriginalYear

		if albumYear != 0 && yearInTitle != albumYear {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Year in title [%d] doesn't match album year (%d)", yearInTitle, albumYear),
			})
		}
	}

	// Check for format indicator (optional but recommended)
	formatPattern := regexp.MustCompile(`\[(FLAC|MP3|AAC|ALAC|WAV|APE|WV)\]`)
	if !formatPattern.MatchString(albumTitle) {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelInfo,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album title could include format indicator [FLAC], [MP3], etc. (optional)",
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
