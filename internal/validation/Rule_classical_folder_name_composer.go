package validation

import (
	"fmt"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// ComposerInFolderName checks that folder name contains composer (classical.folder_name)
// The album title should include the primary composer's name
func (r *Rules) ComposerInFolderName(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "classical.folder_name",
		name:   "Folder name should contain composer name",
		level:  domain.LevelWarning,
		weight: 0.5,
	}
	
	var issues []domain.ValidationIssue
	
	albumTitle := actual.Title()
	if albumTitle == "" {
		return meta.Pass()
	}
	
	tracks := actual.Tracks()
	if len(tracks) == 0 {
		return meta.Pass()
	}
	
	// Find the primary composer(s) across all tracks
	composerCounts := make(map[string]int)
	composerFullNames := make(map[string]string)
	
	for _, track := range tracks {
		for _, artist := range track.Artists() {
			if artist.Role() == domain.RoleComposer {
				name := artist.Name()
				lastName := extractPrimaryLastName(name)
				composerCounts[lastName]++
				composerFullNames[lastName] = name
			}
		}
	}
	
	if len(composerCounts) == 0 {
		// No composers found - will be caught by other rules
		return meta.Pass()
	}
	
	// Find the most frequent composer
	var primaryComposer string
	maxCount := 0
	for lastName, count := range composerCounts {
		if count > maxCount {
			maxCount = count
			primaryComposer = lastName
		}
	}
	
	// Check if composer name appears in album title
	if primaryComposer == "" {
		return meta.Pass()
	}
	
	albumTitleLower := strings.ToLower(albumTitle)
	composerLower := strings.ToLower(primaryComposer)
	fullNameLower := strings.ToLower(composerFullNames[primaryComposer])
	
	// Check for composer mention
	if !strings.Contains(albumTitleLower, composerLower) && !strings.Contains(albumTitleLower, fullNameLower) {
		// Check if this is a "Various Artists" compilation
		if strings.Contains(albumTitleLower, "various") {
			return meta.Pass() // Various artist compilations don't need composer in name
		}
		
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			0,
			meta.id,
			fmt.Sprintf("Album title '%s' should include primary composer name '%s'",
				albumTitle, composerFullNames[primaryComposer]),
		))
	}
	
	// Additional check: if composer is mentioned, prefer full name or proper abbreviation
	if strings.Contains(albumTitleLower, composerLower) && !strings.Contains(albumTitleLower, fullNameLower) {
		// Check if it's an acceptable abbreviation
		fullName := composerFullNames[primaryComposer]
		if !isAcceptableAbbreviation(albumTitle, fullName) {
			issues = append(issues, domain.NewIssue(
				domain.LevelInfo,
				0,
				meta.id,
				fmt.Sprintf("Album title contains composer surname '%s', full name '%s' or abbreviation recommended",
					primaryComposer, fullName),
			))
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}
