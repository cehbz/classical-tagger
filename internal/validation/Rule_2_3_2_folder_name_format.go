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
func (r *Rules) FolderNameFormat(actual, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.2",
		Name:   "Folder name format: Artist - Album [Year] [Format]",
		Level:  domain.LevelWarning, // Warning because format varies
		Weight: 0.5,
	}

	if actual == nil || actual.Title == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	albumTitle := actual.Title
	// Check for basic structure: should have " - " separator
	if !strings.Contains(albumTitle, " - ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title '%s' should contain ' - ' separator between artist and title", albumTitle),
		})
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
	// Allow formats like [FLAC], [FLAC 96-24], [MP3 V0], etc.
	formatPattern := regexp.MustCompile(`\[(FLAC|MP3|AAC|ALAC|WAV|APE|WV)(\s+[^\]]+)?\]`)
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
