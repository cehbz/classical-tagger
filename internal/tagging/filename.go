package tagging

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// GenerateFilename generates a filename for a track following rule 2.3.13 format.
// Format: "## - Track Title.flac"
// Uses leading zeros if totalTracks > 9 (rule 2.3.14).
func GenerateFilename(track *domain.Track, totalTracks int) string {
	// Determine track number format
	trackNumStr := fmt.Sprintf("%d", track.Track)
	if totalTracks > 9 {
		trackNumStr = fmt.Sprintf("%02d", track.Track)
	}

	// Sanitize title for filename
	sanitizedTitle := SanitizeFilename(track.Title)
	if sanitizedTitle == "" {
		sanitizedTitle = "Untitled"
	}

	// Format: "## - Title.flac"
	filename := fmt.Sprintf("%s - %s.flac", trackNumStr, sanitizedTitle)

	return filename
}

// SanitizeFilename sanitizes a string for use as a filename.
// Removes invalid filesystem characters and handles edge cases.
func SanitizeFilename(name string) string {
	if name == "" {
		return ""
	}

	// Remove invalid filesystem characters: / \ : * ? " < > |
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = invalidChars.ReplaceAllString(name, "")

	// Remove leading/trailing spaces and dots (Windows doesn't allow trailing dots)
	name = strings.Trim(name, " .")

	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	name = spacePattern.ReplaceAllString(name, " ")

	// Windows reserved names (CON, PRN, AUX, NUL, COM1-9, LPT1-9)
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

	// Ensure filename doesn't exceed reasonable length
	// Leave room for track number prefix and extension
	maxLength := 170 // Reasonable limit, will be trimmed further if needed for path
	if len(name) > maxLength {
		name = name[:maxLength]
		name = strings.TrimRight(name, " .")
	}

	return name
}

// GenerateDiscSubdirectoryName generates a subdirectory name for a disc.
// Format: "Disc N" or disc title if available.
func GenerateDiscSubdirectoryName(discNum int, discTitle string) string {
	if discTitle != "" {
		sanitized := SanitizeDirectoryName(discTitle)
		if sanitized != "" {
			return sanitized
		}
	}
	return fmt.Sprintf("Disc %d", discNum)
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
