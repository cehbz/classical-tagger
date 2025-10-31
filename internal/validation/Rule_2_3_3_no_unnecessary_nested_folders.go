package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// NoUnnecessaryNestedFolders checks that there are no extra nested folders (rule 2.3.3)
// Acceptable: CD1/, CD2/, Disc1/, etc. for multi-disc releases
// Not acceptable: Artist/Album/CD1/files or Album/Year/files
func (r *Rules) NoUnnecessaryNestedFolders(actualTrack, _ *domain.Track, actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.3",
		Name:   "No unnecessary nested folders beyond disc folders",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	if actualTrack == nil || actualTrack.Name == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	var issues []domain.ValidationIssue

	// Check each track's path
	fileName := actualTrack.Name
	if fileName == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Split path into components
	pathParts := strings.Split(fileName, "/")

	// If there's only a filename (no path), that's fine
	if len(pathParts) <= 1 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check nesting depth
	folderCount := len(pathParts) - 1 // Subtract 1 for the filename itself

	// Determine if album is multi-disc
	isMultiDisc := actualAlbum != nil && actualAlbum.IsMultiDisc()

	// Rule 2.3.3 logic:
	// - Single disc album: Reject ALL folders (depth > 0)
	// - Multi-disc album: Reject non-disc folders AND reject depth > 1 (only disc folders allowed at depth 1)

	if folderCount == 0 {
		// No folders - always acceptable
		return RuleResult{Meta: meta, Issues: nil}
	}

	if !isMultiDisc {
		// Single disc album: no folders allowed
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Unnecessary folder nesting in path '%s' (single disc albums must have all files in main folder)",
				formatTrackNumber(actualTrack), fileName),
		})
		return RuleResult{Meta: meta, Issues: issues}
	}

	// Multi-disc album: check folder structure
	if folderCount > 1 {
		// Depth > 1 is never allowed
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s: Unnecessary folder nesting in path '%s' (depth > 1 not allowed)",
				formatTrackNumber(actualTrack), fileName),
		})
		return RuleResult{Meta: meta, Issues: issues}
	}

	// folderCount == 1 for multi-disc: must be a disc folder
	if folderCount == 1 {
		folder := pathParts[0]
		if !isDiscFolder(folder) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Unnecessary folder nesting in path '%s' (only disc folders like CD1/, Disc2/ allowed for multi-disc albums)",
					formatTrackNumber(actualTrack), fileName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// isDiscFolder checks if a folder name is a disc-related folder
// Accepts: CD1, CD2, Disc1, Disc2, Disk1, etc.
func isDiscFolder(folderName string) bool {
	lower := strings.ToLower(folderName)

	// Check for common disc folder patterns
	prefixes := []string{"cd", "disc", "disk", "dvd"}

	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			// Check if followed by number or just the prefix
			rest := lower[len(prefix):]
			if rest == "" {
				return true
			}
			// Check if rest is a number
			for _, r := range rest {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
	}

	return false
}
