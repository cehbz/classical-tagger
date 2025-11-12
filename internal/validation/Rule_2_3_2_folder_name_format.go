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

// FolderNameFormat checks that directory name follows proper folder naming format (rule 2.3.2)
// Format: "Artist - Album [Year] [Format] [Extra Info]" or variations
func (r *Rules) FolderNameFormat(actual, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.2",
		Name:   "Folder name format: Artist - Album [Year] [Format]",
		Level:  domain.LevelWarning, // Warning because format varies
		Weight: 0.5,
	}

	if actual == nil || actual.RootPath == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Extract directory name from root_path (last component)
	dirName := actual.RootPath
	if idx := strings.LastIndex(dirName, "/"); idx >= 0 {
		dirName = dirName[idx+1:]
	}
	if idx := strings.LastIndex(dirName, "\\"); idx >= 0 {
		dirName = dirName[idx+1:]
	}

	// Check for basic structure: should have " - " separator
	if !strings.Contains(dirName, " - ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album: 2.3.2 - Album title '%s' should contain ' - ' separator between artist and title", dirName),
		})
	}

	// Check for year presence (non-prescriptive about format)
	// Accept year in various formats: [YYYY], (YYYY), - YYYY, etc.
	yearPattern := regexp.MustCompile(`(\[(\d{4})\]|\((\d{4})\)|-\s*(\d{4})(\s|\[|$)|(\d{4}))`)
	yearMatches := yearPattern.FindStringSubmatch(dirName)

	if len(yearMatches) == 0 {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelWarning,
			Track:   0,
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album: 2.3.2 - Album title '%s' should include year in brackets [YYYY]", dirName),
		})
	} else {
		// Extract year from whichever format matched
		// Groups: 0=full match, 1=full match, 2=[YYYY], 3=(YYYY), 4=- YYYY, 5=separator, 6=standalone
		var yearInDir int
		yearGroups := []int{2, 3, 4, 6} // Skip group 5 (separator)
		for _, i := range yearGroups {
			if i < len(yearMatches) && yearMatches[i] != "" {
				yearInDir, _ = strconv.Atoi(yearMatches[i])
				break
			}
		}
		albumYear := actual.OriginalYear

		// Verify year matches album year if available
		if albumYear != 0 && yearInDir != 0 && yearInDir != albumYear {
			issues = append(issues, domain.ValidationIssue{
				Level:   domain.LevelWarning,
				Track:   0,
				Rule:    meta.ID,
				Message: fmt.Sprintf("Year in directory name (%d) doesn't match album year (%d)", yearInDir, albumYear),
			})
		}
	}

	// Check for format indicator (optional but recommended)
	// Allow formats like [FLAC], [FLAC 96-24], [MP3 V0], etc.
	formatPattern := regexp.MustCompile(`\[(FLAC|MP3|AAC|ALAC|WAV|APE|WV)(\s+[^\]]+)?\]`)
	if !formatPattern.MatchString(dirName) {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelInfo,
			Track:   0,
			Rule:    meta.ID,
			Message: "Album: 2.3.2 - Album title could include format indicator [FLAC], [MP3], etc. (optional)",
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}
