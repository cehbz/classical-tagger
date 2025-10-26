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
		id:     "2.3.2",
		name:   "Folder name format: Artist - Album [Year] [Format]",
		level:  domain.LevelWarning, // Warning because format varies
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	albumTitle := actual.Title()
	if albumTitle == "" {
		// Will be caught by RequiredTags
		return meta.Pass()
	}
	
	// Check for basic structure: should have " - " separator
	if !strings.Contains(albumTitle, " - ") {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			fmt.Sprintf("Album title '%s' should contain ' - ' separator between artist and title", albumTitle),
		))
		return meta.Fail(issues...)
	}
	
	// Check for year in brackets
	yearPattern := regexp.MustCompile(`\[(\d{4})\]`)
	yearMatches := yearPattern.FindStringSubmatch(albumTitle)
	
	if len(yearMatches) == 0 {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			fmt.Sprintf("Album title '%s' should include year in brackets [YYYY]", albumTitle),
		))
	} else {
		// Verify year matches album year
		yearInTitle, _ := strconv.Atoi(yearMatches[1])
		albumYear := actual.OriginalYear()
		
		if albumYear != 0 && yearInTitle != albumYear {
			issues = append(issues, domain.NewIssue(
				domain.LevelWarning,
				0,
				meta.id,
				fmt.Sprintf("Year in title [%d] doesn't match album year (%d)", yearInTitle, albumYear),
			))
		}
	}
	
	// Check for format indicator (optional but recommended)
	formatPattern := regexp.MustCompile(`\[(FLAC|MP3|AAC|ALAC|WAV|APE|WV)\]`)
	if !formatPattern.MatchString(albumTitle) {
		issues = append(issues, domain.NewIssue(
			domain.LevelInfo,
			0,
			meta.id,
			"Album title could include format indicator [FLAC], [MP3], etc. (optional)",
		))
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
