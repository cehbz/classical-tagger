package tagging

import (
	"fmt"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// GenerateDirectoryName generates a directory name for a torrent following classical music conventions.
// Format: "Composer - Album Title (Performers) - Year [FLAC]"
// Falls back to simpler formats if too long.
// Minimum: "Album Title" (rule 2.3.2)
func GenerateDirectoryName(torrent *domain.Torrent) string {
	tracks := torrent.Tracks()

	// Get album title
	albumTitle := SanitizeDirectoryName(torrent.Title)
	if albumTitle == "" {
		albumTitle = "Untitled Album"
	}
	dirName := albumTitle
	dirNameLen := len(dirName)

	formatIndicator := " [FLAC]"
	if dirNameLen+len(formatIndicator) > 180 {
		return dirName
	}
	dirNameLen += len(formatIndicator)
	yearStr := ""
	// Get year
	if torrent.OriginalYear > 0 {
		yearStr = fmt.Sprintf(" - %d", torrent.OriginalYear)
	}
	if dirNameLen+len(yearStr) > 180 {
		return dirName + formatIndicator
	}
	dirNameLen += len(yearStr)

	// Get primary composer(s)
	composers := getPrimaryComposers(tracks)
	composerStr := formatComposersForDirectory(composers) + " - "
	if dirNameLen+len(composerStr) > 180 {
		return dirName + yearStr + formatIndicator
	}
	dirName = composerStr + dirName
	dirNameLen += len(composerStr)

	// Get performers (for optional inclusion)
	performers := getPrimaryPerformers(tracks)
	performerStr := ""
	if len(performers) > 0 {
		performerStr = " (" + formatPerformersForDirectory(performers) + ")"
	}
	if dirNameLen+len(performerStr) > 180 {
		return dirName + yearStr + formatIndicator
	}
	return dirName + performerStr + yearStr + formatIndicator
}

// getPrimaryComposers extracts the primary composer(s) from tracks.
// Returns the most frequent composer, or first composer if tied.
func getPrimaryComposers(tracks []*domain.Track) []string {
	composerCounts := make(map[string]int)
	composerOrder := make([]string, 0)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role == domain.RoleComposer && artist.Name != "" {
				if composerCounts[artist.Name] == 0 {
					composerOrder = append(composerOrder, artist.Name)
				}
				composerCounts[artist.Name]++
			}
		}
	}

	if len(composerCounts) == 0 {
		return nil
	}

	// Find most frequent composer
	maxCount := 0
	var primaryComposer string
	for name, count := range composerCounts {
		if count > maxCount {
			maxCount = count
			primaryComposer = name
		}
	}

	// Return primary composer, or all if multiple composers appear frequently
	var result []string
	if maxCount > len(tracks)/2 {
		// Single dominant composer
		result = []string{primaryComposer}
	} else {
		// Multiple composers - return all
		result = composerOrder
	}

	return result
}

// formatComposersForDirectory formats composer names for directory name.
// Uses last names if multiple composers to save space.
func formatComposersForDirectory(composers []string) string {
	if len(composers) == 0 {
		return ""
	}

	if len(composers) == 1 {
		// Single composer - use full name or last name
		name := composers[0]
		parts := strings.Fields(name)
		if len(parts) > 1 {
			// Use last name for brevity
			return parts[len(parts)-1]
		}
		return name
	}

	// Multiple composers - use last names separated by commas
	var lastNames []string
	for _, composer := range composers {
		parts := strings.Fields(composer)
		if len(parts) > 0 {
			lastNames = append(lastNames, parts[len(parts)-1])
		} else {
			lastNames = append(lastNames, composer)
		}
	}
	return strings.Join(lastNames, ", ")
}

// getPrimaryPerformers extracts primary performers (non-composers) from tracks.
// Returns performers that appear in most tracks.
func getPrimaryPerformers(tracks []*domain.Track) []string {
	performerCounts := make(map[string]int)
	performerOrder := make([]string, 0)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role != domain.RoleComposer && artist.Name != "" {
				if performerCounts[artist.Name] == 0 {
					performerOrder = append(performerOrder, artist.Name)
				}
				performerCounts[artist.Name]++
			}
		}
	}

	if len(performerCounts) == 0 {
		return nil
	}

	// Return performers that appear in at least 50% of tracks
	var result []string
	threshold := len(tracks) / 2
	if threshold == 0 {
		threshold = 1
	}

	for _, name := range performerOrder {
		if performerCounts[name] >= threshold {
			result = append(result, name)
		}
	}

	// Limit to 3 performers to keep directory name reasonable
	if len(result) > 3 {
		result = result[:3]
	}

	return result
}

// formatPerformersForDirectory formats performer names for directory name.
// Uses abbreviated names if needed.
func formatPerformersForDirectory(performers []string) string {
	if len(performers) == 0 {
		return ""
	}

	// Use last names or abbreviated names
	var formatted []string
	for _, performer := range performers {
		parts := strings.Fields(performer)
		if len(parts) > 1 {
			// Use last name
			formatted = append(formatted, parts[len(parts)-1])
		} else {
			formatted = append(formatted, performer)
		}
	}

	result := strings.Join(formatted, ", ")

	// Limit length
	if len(result) > 50 {
		result = result[:47] + "..."
	}

	return result
}
