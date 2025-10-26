package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// NoUnnecessaryNestedFolders checks that there are no extra nested folders (rule 2.3.3)
// Acceptable: CD1/, CD2/, Disc1/, etc. for multi-disc releases
// Not acceptable: Artist/Album/CD1/files or Album/Year/files
func (r *Rules) NoUnnecessaryNestedFolders(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.3",
		name:   "No unnecessary nested folders beyond disc folders",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	
	// Check each track's path
	for _, track := range actual.Tracks() {
		fileName := track.Name()
		if fileName == "" {
			continue
		}
		
		// Split path into components
		pathParts := strings.Split(fileName, "/")
		
		// If there's only a filename (no path), that's fine
		if len(pathParts) <= 1 {
			continue
		}
		
		// Check nesting depth
		folderCount := len(pathParts) - 1 // Subtract 1 for the filename itself
		
		// Acceptable patterns:
		// - Single level: "CD1/01 - Track.flac" (1 folder)
		// - No folders: "01 - Track.flac" (0 folders)
		// Unacceptable:
		// - Two or more levels: "Artist/Album/01 - Track.flac" (2+ folders)
		// - Exception: "CD1/CD1-01/01 - Track.flac" might be acceptable for complex releases
		
		if folderCount > 1 {
			// Check if all folders are disc-related
			allDiscFolders := true
			for i := 0; i < folderCount; i++ {
				folder := pathParts[i]
				if !isDiscFolder(folder) {
					allDiscFolders = false
					break
				}
			}
			
			if !allDiscFolders {
				issues = append(issues, domain.NewIssue(
					domain.LevelError,
					track.Track(),
					meta.id,
					fmt.Sprintf("Track %s: Unnecessary folder nesting in path '%s' (only disc folders like CD1/, Disc2/ allowed)",
						formatTrackNumber(track), fileName),
				))
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
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
