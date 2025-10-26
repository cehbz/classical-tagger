package validation

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// MultiDiscFolderSorting checks multi-disc folder naming for proper sorting (rule 2.3.19)
// INFO level - suggests folder names that sort properly
func (r *Rules) MultiDiscFolderSorting(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.19",
		name:   "Multi-disc folders should sort properly (CD1, CD2, not CD1, CD10, CD2)",
		level:  domain.LevelInfo,
		weight: 0.1,
	}
	
	var issues []domain.ValidationIssue
	
	// Find disc count
	maxDisc := 0
	for _, track := range actual.Tracks() {
		if track.Disc() > maxDisc {
			maxDisc = track.Disc()
		}
	}
	
	// Only relevant for multi-disc releases
	if maxDisc <= 1 {
		return meta.Pass()
	}
	
	// Collect folder names used
	discFolders := make(map[int]string)
	for _, track := range actual.Tracks() {
		if track.Name() == "" {
			continue
		}
		
		// Extract folder name if present
		dir := filepath.Dir(track.Name())
		if dir != "." && dir != "" {
			discFolders[track.Disc()] = dir
		}
	}
	
	if len(discFolders) == 0 {
		return meta.Pass() // No folders used
	}
	
	// Check if folder names will sort properly
	var folderNames []string
	for _, folder := range discFolders {
		folderNames = append(folderNames, folder)
	}
	
	// Check natural sort order
	sortedFolders := make([]string, len(folderNames))
	copy(sortedFolders, folderNames)
	sort.Strings(sortedFolders)
	
	// Check if numeric order matches alphabetic sort
	if maxDisc >= 10 {
		// Potential issue if not zero-padded
		for disc, folder := range discFolders {
			// Check if single-digit discs are zero-padded
			if disc < 10 {
				lowerFolder := strings.ToLower(folder)
				// Common patterns: CD1, CD01, Disc1, Disc01
				hasProperPadding := strings.Contains(lowerFolder, "cd0") ||
					strings.Contains(lowerFolder, "disc0") ||
					strings.Contains(lowerFolder, "disk0")
				
				if !hasProperPadding && (strings.Contains(lowerFolder, "cd") ||
					strings.Contains(lowerFolder, "disc") ||
					strings.Contains(lowerFolder, "disk")) {
					issues = append(issues, domain.NewIssue(
						domain.LevelInfo,
						0,
						meta.id,
						fmt.Sprintf("Disc %d folder '%s' may not sort properly (recommend CD01, CD02, etc. for 10+ discs)",
							disc, folder),
					))
				}
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
