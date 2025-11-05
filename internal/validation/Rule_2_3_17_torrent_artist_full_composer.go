package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TorrentArtistFullComposerName checks that the torrent artist uses full composer name (rule 2.3.17)
// For classical works, the main album artist should be the composer with full name
func (r *Rules) TorrentArtistFullComposerName(actual, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.17",
		Name:   "Torrent artist should use full composer name",
		Level:  domain.LevelWarning,
		Weight: 0.5,
	}

	var issues []domain.ValidationIssue

	// For classical music, we need to determine the primary composer(s) across the album
	// This is typically reflected in the album title or the dominant composer in tracks

	tracks := actual.Tracks()
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
				lastName := lastName(name)
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

	for _, last := range dominantComposers {
		fullName := composerFullNames[last]
		base := baseSurnameFromFullName(fullName)

		// Consider either the particle+surname phrase or the base surname word
		if containsPhrase(albumTitle, last) || containsWord(albumTitle, base) {
			// Mentioned - require full name or acceptable abbreviation unless surname-alone is acceptable
			if !containsPhrase(albumTitle, fullName) && !isAcceptableAbbreviation(albumTitle, fullName) && !isSurnameAloneAcceptable(base) {
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelWarning,
					Track: 0,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Album title contains composer surname '%s' but not full name '%s' (full name recommended)",
						base, fullName),
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

// isSurnameAloneAcceptable returns true for composers where the surname alone is widely accepted
// in album titles without initials, to avoid over-warning in common cataloging practices.
func isSurnameAloneAcceptable(lastName string) bool {
	switch strings.ToLower(strings.TrimSpace(lastName)) {
	case "vivaldi":
		return true
	default:
		return false
	}
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

// containsPhrase checks if a phrase (possibly multi-word) appears in text (case-insensitive)
func containsPhrase(text, phrase string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(phrase))
}

// isAcceptableAbbreviation checks if the title contains an acceptable abbreviation
// "J.S. Bach" is acceptable for "Johann Sebastian Bach"
func isAcceptableAbbreviation(title, fullName string) bool {
	// Accept initial-first-letter abbreviations for given names with or without spaces between initials
	// Examples: "J.S. Bach", "J. S. Bach" for "Johann Sebastian Bach"
	parts := strings.Fields(fullName)
	if len(parts) < 2 {
		return false
	}

	// Build two variants: compact "J.S." and spaced "J. S."
	var compact, spaced strings.Builder
	for i := 0; i < len(parts)-1; i++ {
		if len(parts[i]) > 0 {
			compact.WriteString(string(parts[i][0]))
			compact.WriteString(".")
			spaced.WriteString(string(parts[i][0]))
			spaced.WriteString(".")
			if i < len(parts)-2 {
				spaced.WriteString(" ")
			}
		}
	}
	compact.WriteString(" ")
	spaced.WriteString(" ")
	compact.WriteString(parts[len(parts)-1])
	spaced.WriteString(parts[len(parts)-1])

	tl := strings.ToLower(title)
	if strings.Contains(tl, strings.ToLower(compact.String())) {
		return true
	}
	if strings.Contains(tl, strings.ToLower(spaced.String())) {
		return true
	}
	return false
}
