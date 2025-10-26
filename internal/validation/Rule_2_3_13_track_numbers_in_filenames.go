package validation

import (
	"fmt"
	"regexp"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// trackNumberPattern matches track numbers at start of filename
// Accepts: "01", "1", "01.", "1.", "01 -", "1 -", "01_", etc.
var trackNumberPattern = regexp.MustCompile(`^(\d+)[\s\-_\.]+`)

// TrackNumbersInFilenames checks that all filenames contain track numbers (rule 2.3.13)
// Exception: Single-track torrents don't require track numbers
func (r *Rules) TrackNumbersInFilenames(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.13",
		name:   "Track numbers required in file names",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	tracks := actual.Tracks()
	
	// Exception: Single-track torrents don't require track numbers
	if len(tracks) == 1 {
		return meta.Pass()
	}
	
	for _, track := range tracks {
		fileName := track.Name()
		if fileName == "" {
			// No filename set - this will be caught by other rules
			continue
		}
		
		// Extract just the filename (not path) for checking
		// In case Name() contains path like "CD1/01 - Track.flac"
		parts := strings.Split(fileName, "/")
		justFileName := parts[len(parts)-1]
		
		// Check if filename starts with a track number
		matches := trackNumberPattern.FindStringSubmatch(justFileName)
		if len(matches) == 0 {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				track.Track(),
				meta.id,
				fmt.Sprintf("Filename missing track number: %s", fileName),
			))
			continue
		}
		
		// Optional: Verify the extracted number matches the track number
		// This is a bonus check beyond the rule requirement
		// (commented out as rule 2.3.13 only requires presence, not correctness)
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
