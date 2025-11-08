package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// NoCombinedTags checks that tags don't combine different types of information (rule 2.3.18.3)
// Tags should not combine different field types (e.g., track number and title in title tag)
func (r *Rules) NoCombinedTags(actualTrack, _ *domain.Track, actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.18.3",
		Name:   "No combined tags - each tag should contain only one type of information",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	title := actualTrack.Title

	// Check for multiple works in title (should be separate tracks)
	// Pattern: "Work 1 / Work 2" or "Work 1; Work 2"
	for _, sep := range []string{" / ", "; ", " & ", ", ", " and "} {
		if strings.Contains(title, sep) {
			// Check if this looks like multiple works
			parts := strings.Split(title, sep)
			if len(parts) >= 2 && len(parts[0]) > 10 && len(parts[1]) > 10 {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelInfo,
					Track: actualTrack.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s: Title may contain multiple works '%s' (consider separate tracks)",
						formatTrackNumber(actualTrack), title),
				})
				break
			}
		}
	}

	// Check if title contains track number (combining track number with title)
	// Patterns: "01 - Title", "Track 1: Title", "1. Title", etc.
	// But NOT "Symphony No. 5" (that's part of the work title)
	trackNumPattern := regexp.MustCompile(`(?i)^\s*(\d{1,3})[\s\-._:]+|^\s*track\s*(\d{1,3})[\s\-._:]+`)
	if matches := trackNumPattern.FindStringSubmatch(title); len(matches) > 0 {
		// Check if the matched number matches the actual track number
		matchedNum := ""
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				matchedNum = matches[i]
				break
			}
		}
		if matchedNum != "" {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelWarning,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Title contains track number '%s' (track number should be in separate tag)",
					formatTrackNumber(actualTrack), matchedNum),
			})
		}
	}

	// Check album title for disc number without meaningful subtitle
	if actualTorrent != nil {
		albumTitle := actualTorrent.Title
		// Pattern: "[Album Name] Disc n" or "[Album Name] (Disc n)" without meaningful subtitle
		discPattern := regexp.MustCompile(`(?i)\s*(?:\(|\[)?\s*disc\s*(\d+)\s*(?:\)|\])?\s*$`)
		if matches := discPattern.FindStringSubmatch(albumTitle); len(matches) > 0 {
			// Check if there's a meaningful subtitle before the disc number
			// A meaningful subtitle should have a separator (dash, colon) or be substantial
			beforeDisc := strings.TrimSpace(discPattern.ReplaceAllString(albumTitle, ""))
			hasSeparator := strings.Contains(beforeDisc, " - ") || strings.Contains(beforeDisc, ": ")
			isSubstantial := len(beforeDisc) > 10
			
			if beforeDisc == "" || (!hasSeparator && !isSubstantial) {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelWarning,
					Track: 0, // Album-level issue
					Rule:  meta.ID,
					Message: fmt.Sprintf("Album title contains disc number without meaningful subtitle: '%s'",
						albumTitle),
				})
			}
		}
	}

	return RuleResult{Meta: meta, Issues: issues}
}

