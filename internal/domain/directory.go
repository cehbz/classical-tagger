package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// GenerateDirectoryName generates a directory name for a torrent following classical music conventions.
// Format: "Composer - Album Title (Performers) - Year [FLAC]"
// Falls back to simpler formats if too long.
// Minimum: "Album Title" (rule 2.3.2)
func GenerateDirectoryName(torrent *Torrent) string {
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
// Returns the most frequent composer, or all composers if no single composer appears on more than half the tracks.
func getPrimaryComposers(tracks []*Track) []string {
	composerCounts := make(map[string]int)
	composerOrder := make([]string, 0)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role == RoleComposer && artist.Name != "" {
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

	// Return primary composer only if they appear on more than half the tracks
	// Otherwise return all composers
	var result []string
	if maxCount > len(tracks)/2 {
		// Single dominant composer (>50% of tracks)
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
func getPrimaryPerformers(tracks []*Track) []string {
	performerCounts := make(map[string]int)
	performerOrder := make([]string, 0)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if artist.Role != RoleComposer && artist.Name != "" {
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

// SanitizeDirectoryName sanitizes a string for use as a directory name.
// Similar to filename sanitization but allows some additional characters.
func SanitizeDirectoryName(name string) string {
	if name == "" {
		return ""
	}

	// Remove invalid filesystem characters: / \ : * ? " < > |
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = invalidChars.ReplaceAllString(name, "")

	// Remove leading/trailing spaces and dots
	name = strings.Trim(name, " .")

	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	name = spacePattern.ReplaceAllString(name, " ")

	// Windows reserved names
	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}
	upperName := strings.ToUpper(name)
	if reservedNames[upperName] || strings.HasPrefix(upperName, "COM") || strings.HasPrefix(upperName, "LPT") {
		name = "_" + name
	}

	return name
}
