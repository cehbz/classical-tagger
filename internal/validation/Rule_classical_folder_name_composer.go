package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// ComposerInFolderName checks that folder name contains composer (classical.folder_name)
// The album title should include the primary composer's name
func (r *Rules) ComposerInFolderName(actual, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "classical.folder_name",
		Name:   "Folder name should contain composer name",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	albumTitle := actual.Title
	if albumTitle == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	tracks := actual.Tracks
	if len(tracks) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Find the primary composer(s) across all tracks
	composerCounts := make(map[string]int)
	composerFullNames := make(map[string]string)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleComposer {
				name := artist.Name
				lastName := lastName(name)
				composerCounts[lastName]++
				composerFullNames[lastName] = name
			}
		}
	}

	if len(composerCounts) == 0 {
		// No composers found - will be caught by other rules
		return RuleResult{Meta: meta, Issues: nil}
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
		return RuleResult{Meta: meta, Issues: nil}
	}

	albumTitleLower := strings.ToLower(albumTitle)
	composerLower := strings.ToLower(primaryComposer)
	fullNameStr := composerFullNames[primaryComposer]
	fullNameLower := strings.ToLower(fullNameStr)
	base := strings.ToLower(composerBaseSurname(fullNameStr))

	// Check for composer mention
	if !strings.Contains(albumTitleLower, composerLower) && !strings.Contains(albumTitleLower, fullNameLower) && !strings.Contains(albumTitleLower, base) {
		// Check if this is a "Various Artists" compilation
		if strings.Contains(albumTitleLower, "various") {
			return RuleResult{Meta: meta, Issues: nil} // Various artist compilations don't need composer in name
		}

		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelWarning,
			Track: 0,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Album title '%s' should include primary composer name '%s'",
				albumTitle, composerFullNames[primaryComposer]),
		})
	}

	// Additional check: if composer is mentioned, prefer full name or proper abbreviation
	if (strings.Contains(albumTitleLower, composerLower) || strings.Contains(albumTitleLower, base)) && !strings.Contains(albumTitleLower, fullNameLower) {
		// Check if it's an acceptable abbreviation
		fullName := composerFullNames[primaryComposer]
		if !isAcceptableAbbreviation(albumTitle, fullName) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelInfo,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title contains composer surname '%s', full name '%s' or abbreviation recommended",
					primaryComposer, fullName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

func composerBaseSurname(fullName string) string {
	if strings.Contains(fullName, ",") {
		parts := strings.Split(fullName, ",")
		return strings.TrimSpace(parts[0])
	}
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return fullName
	}
	return parts[len(parts)-1]
}
