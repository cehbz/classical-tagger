package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TorrentArtistFullComposerName checks that the torrent artist uses full composer name (rule 2.3.17)
// For classical works, the main album artist should be the composer with full name
func (r *Rules) TorrentArtistFullComposerName(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.17",
		Name:   "Torrent artist should use full composer name",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	// For classical music, we need to determine the primary composer(s) across the album
	// This is typically reflected in the album title or the dominant composer in tracks

	tracks := actual.Tracks
	if len(tracks) == 0 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Count composers across all tracks
	composerCounts := make(map[string]int)
	composerFullNames := make(map[string]string) // last name -> full name

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleComposer {
				name := artist.Name
				lastName := extractPrimaryLastName(name)
				composerCounts[lastName]++
				composerFullNames[lastName] = name
			}
		}
	}

	// Find the dominant composer(s) - appearing in most tracks
	var dominantComposers []string
	maxCount := 0
	for lastName, count := range composerCounts {
		if count > maxCount {
			maxCount = count
			dominantComposers = []string{lastName}
		} else if count == maxCount {
			dominantComposers = append(dominantComposers, lastName)
		}
	}

	if len(dominantComposers) == 0 {
		// No composers found - will be caught by other rules
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Check album title for composer mentions
	// The "torrent artist" in classical music is typically the composer in the folder/album name
	albumTitle := actual.Title

	for _, lastName := range dominantComposers {
		fullName := composerFullNames[lastName]

		// Check if the last name appears in the album title
		if containsWord(albumTitle, lastName) {
			// Last name is mentioned - check if it's the full name or abbreviated
			if !containsWord(albumTitle, fullName) && !isAcceptableAbbreviation(albumTitle, fullName) {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelWarning,
					Track: 0,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Album title contains composer surname '%s' but not full name '%s' (full name recommended)",
						lastName, fullName),
				})
			}
		} else {
			// Composer not mentioned in album title at all
			// This is INFO level - it's recommended but not required
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelInfo,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Album title does not mention primary composer '%s' (recommended for classical albums)",
					fullName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// extractPrimaryLastName gets the main last name from a composer name
// "Johann Sebastian Bach" -> "Bach"
// "Ludwig van Beethoven" -> "Beethoven" (not "van Beethoven" for this comparison)
func extractPrimaryLastName(composerName string) string {
	lastNames := extractLastNames(composerName)
	if len(lastNames) == 0 {
		return composerName
	}

	// Get the final word of the last name (skip particles like "van", "von")
	lastName := lastNames[0]
	parts := strings.Fields(lastName)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return lastName
}

// containsWord checks if a word appears in text (case-insensitive, word boundary)
func containsWord(text, word string) bool {
	textLower := strings.ToLower(text)
	wordLower := strings.ToLower(word)

	// Simple word boundary check
	words := strings.FieldsFunc(textLower, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-')
	})

	for _, w := range words {
		if w == wordLower || strings.HasPrefix(w, wordLower+".") {
			return true
		}
	}
	return false
}

// isAcceptableAbbreviation checks if the title contains an acceptable abbreviation
// "J.S. Bach" is acceptable for "Johann Sebastian Bach"
func isAcceptableAbbreviation(title, fullName string) bool {
	// Check for common abbreviations: "J.S. Bach", "W.A. Mozart"
	parts := strings.Fields(fullName)
	if len(parts) < 2 {
		return false
	}

	// Build potential abbreviation: "J.S. Bach" for "Johann Sebastian Bach"
	var abbrev strings.Builder
	for i := 0; i < len(parts)-1; i++ {
		if len(parts[i]) > 0 {
			abbrev.WriteString(string(parts[i][0]))
			abbrev.WriteString(".")
			if i < len(parts)-2 {
				abbrev.WriteString(" ")
			}
		}
	}
	abbrev.WriteString(" ")
	abbrev.WriteString(parts[len(parts)-1])

	return containsWord(title, abbrev.String())
}
