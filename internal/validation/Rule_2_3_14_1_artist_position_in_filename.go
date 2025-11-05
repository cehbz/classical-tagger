package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// artistBeforeTrackPattern matches filenames with artist before track number
// e.g., "Artist Name - 01 - Track.flac"
var artistBeforeTrackPattern = regexp.MustCompile(`^[^0-9]+\s*-\s*(\d+)`)

// ArtistPositionInFilename checks that artist names (if present) come AFTER track numbers (rule 2.3.14.1)
// For multi-artist/composer albums, this ensures proper sorting
func (r *Rules) ArtistPositionInFilename(actualTrack, _ *domain.Track, actualTorrent, _ *domain.Torrent) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.14.1",
		Name:   "Artist name must come after track number in filename",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// First, determine if this is a multi-artist/multi-composer album
	isMultiArtist := isMultiComposerAlbum(actualTorrent)

	if !isMultiArtist {
		// For single-artist albums, artist in filename is optional and position doesn't matter
		return RuleResult{Meta: meta, Issues: nil}
	}

	// For multi-artist albums, check each track
	fileName := actualTrack.File.Path
	if fileName == "" {
		return RuleResult{Meta: meta, Issues: nil}
	}

	// Extract just the filename (not path)
	justFileName := filepath.Base(fileName)

	// Check if an artist name appears before the track number
	if artistBeforeTrackPattern.MatchString(justFileName) {
		// Check if this is actually an artist name by comparing with track artists
		if containsArtistName(justFileName, actualTrack.Artists) {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s: Artist name appears before track number in filename '%s' (should be: '01 - Artist - Title.flac')",
					formatTrackNumber(actualTrack), justFileName),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// isMultiComposerAlbum determines if an album has multiple composers
func isMultiComposerAlbum(album *domain.Torrent) bool {
	composers := make(map[string]bool)

	for _, track := range album.Tracks() {
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleComposer {
				composers[artist.Name] = true
			}
		}
	}

	return len(composers) > 1
}

// containsArtistName checks if any artist name appears in the filename prefix
func containsArtistName(filename string, artists []domain.Artist) bool {
	filenameLower := strings.ToLower(filename)

	for _, artist := range artists {
		// Check for last name or full name
		name := artist.Name
		nameLower := strings.ToLower(name)

		// Extract last name
		nameParts := strings.Fields(name)
		var lastName string
		if len(nameParts) > 0 {
			lastName = nameParts[len(nameParts)-1]
		}

		// Check if full name or last name appears before first digit
		firstDigitPos := strings.IndexFunc(filenameLower, func(r rune) bool {
			return r >= '0' && r <= '9'
		})

		if firstDigitPos > 0 {
			prefix := filenameLower[:firstDigitPos]

			// Check for name in prefix
			if strings.Contains(prefix, nameLower) {
				return true
			}
			if lastName != "" && strings.Contains(prefix, strings.ToLower(lastName)) {
				return true
			}
		}
	}

	return false
}
