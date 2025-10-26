package validation

import (
	"fmt"
	"regexp"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// artistBeforeTrackPattern matches filenames with artist before track number
// e.g., "Artist Name - 01 - Track.flac"
var artistBeforeTrackPattern = regexp.MustCompile(`^[^0-9]+\s*-\s*(\d+)`)

// ArtistPositionInFilename checks that artist names (if present) come AFTER track numbers (rule 2.3.14.1)
// For multi-artist/composer albums, this ensures proper sorting
func (r *Rules) ArtistPositionInFilename(actual, reference *domain.Album) RuleResult {
	meta := RuleMetadata{
		id:     "2.3.14.1",
		name:   "Artist name must come after track number in filename",
		level:  domain.LevelError,
		weight: 1.0,
	}
	
	var issues []domain.ValidationIssue
	
	// First, determine if this is a multi-artist/multi-composer album
	isMultiArtist := isMultiComposerAlbum(actual)
	
	if !isMultiArtist {
		// For single-artist albums, artist in filename is optional and position doesn't matter
		return meta.Pass()
	}
	
	// For multi-artist albums, check each track
	for _, track := range actual.Tracks() {
		fileName := track.Name()
		if fileName == "" {
			continue
		}
		
		// Extract just the filename (not path)
		parts := strings.Split(fileName, "/")
		justFileName := parts[len(parts)-1]
		
		// Check if an artist name appears before the track number
		if artistBeforeTrackPattern.MatchString(justFileName) {
			// Check if this is actually an artist name by comparing with track artists
			if containsArtistName(justFileName, track.Artists()) {
				issues = append(issues, domain.NewIssue(
					domain.LevelError,
					track.Track(),
					meta.id,
					fmt.Sprintf("Track %s: Artist name appears before track number in filename '%s' (should be: '01 - Artist - Title.flac')",
						formatTrackNumber(track), justFileName),
				))
			}
		}
	}
	
	if len(issues) == 0 {
		return meta.Pass()
	}
	return meta.Fail(issues...)
}

// isMultiComposerAlbum determines if an album has multiple composers
func isMultiComposerAlbum(album *domain.Album) bool {
	composers := make(map[string]bool)
	
	for _, track := range album.Tracks() {
		for _, artist := range track.Artists() {
			if artist.Role() == domain.RoleComposer {
				composers[artist.Name()] = true
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
		name := artist.Name()
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
