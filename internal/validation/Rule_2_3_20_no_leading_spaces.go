package validation

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumNoLeadingSpaces checks that no file or folder names have leading spaces (album: 2.3.20-album)
func (r *Rules) AlbumNoLeadingSpaces(actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.20-album",
		Name:   "No leading spaces in file or folder names",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Check album title
	if strings.HasPrefix(actualAlbum.Title, " ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0, // Album-level issue
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album title has leading space: '%s'", actualAlbum.Title),
		})
	}

	// Check folder name
	if strings.HasPrefix(actualAlbum.FolderName, " ") {
		issues = append(issues, domain.ValidationIssue{
			Level:   domain.LevelError,
			Track:   0, // Album-level issue
			Rule:    meta.ID,
			Message: fmt.Sprintf("Album folder name has leading space: '%s'", actualAlbum.FolderName),
		})
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// TrackNoLeadingSpaces checks that no file or folder names have leading spaces (track: 2.3.20)
func (r *Rules) TrackNoLeadingSpaces(actualTrack, _ *domain.Track, actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.20",
		Name:   "No leading spaces in file or folder names",
		Level:  domain.LevelError,
		Weight: 1.0,
	}

	var issues []domain.ValidationIssue

	// Track-level rule ignores album-level title spacing to avoid double-counting

	// Check filename
	fileName := actualTrack.Name
	if fileName != "" {
		// Check the base filename and any path components
		parts := strings.Split(fileName, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, " ") {
				var location string
				if i == len(parts)-1 {
					location = "filename"
				} else {
					location = "folder name"
				}
				issues = append(issues, domain.ValidationIssue{
					Level: domain.LevelError,
					Track: actualTrack.Track,
					Rule:  meta.ID,
					Message: fmt.Sprintf("Track %s has leading space in %s: '%s'",
						formatTrackNumber(actualTrack), location, part),
				})
			}
		}
	}

	// Check track title (tag value)
	if strings.HasPrefix(actualTrack.Title, " ") {
		issues = append(issues, domain.ValidationIssue{
			Level: domain.LevelError,
			Track: actualTrack.Track,
			Rule:  meta.ID,
			Message: fmt.Sprintf("Track %s title tag has leading space: '%s'",
				formatTrackNumber(actualTrack), actualTrack.Title),
		})
	}

	// Check artist names
	for _, artist := range actualTrack.Artists {
		if strings.HasPrefix(artist.Name, " ") {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelError,
				Track: actualTrack.Track,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Track %s artist name has leading space: '%s'",
					formatTrackNumber(actualTrack), artist.Name),
			})
		}
	}
	return RuleResult{Meta: meta, Issues: issues}
}

// formatTrackNumber formats a track for error messages
func formatTrackNumber(track *domain.Track) string {
	if track.Disc > 1 {
		return fmt.Sprintf("%d-%d", track.Disc, track.Track)
	}
	return fmt.Sprintf("%d", track.Track)
}
