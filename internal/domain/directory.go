package domain

import (
	"regexp"
	"strings"
)

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
