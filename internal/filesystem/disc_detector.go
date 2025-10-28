package filesystem

import (
	"strings"
)

// IsDiscDirectory checks if a directory name represents a disc folder.
// Accepts: CD1, CD2, cd1, Disc1, Disc 2, Disk1, DVD1, CD, Disc, etc.
// Rejects: Artist, Album, 1963, CDextra (contains non-digits after prefix)
//
// This is the canonical implementation used throughout the codebase.
func IsDiscDirectory(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return false
	}

	// Common disc folder prefixes
	prefixes := []string{"cd", "disc", "disk", "dvd"}

	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			// Get the part after the prefix
			rest := strings.TrimSpace(lower[len(prefix):])
			
			// If nothing after prefix, it's valid (e.g., "CD", "Disc")
			if rest == "" {
				return true
			}
			
			// If there's content after prefix, it must be all digits
			for _, r := range rest {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
	}

	return false
}