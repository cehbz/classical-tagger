package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// DirectoryValidator validates directory and file naming conventions.
type DirectoryValidator struct{}

// NewDirectoryValidator creates a new DirectoryValidator.
func NewDirectoryValidator() *DirectoryValidator {
	return &DirectoryValidator{}
}

// ValidatePath checks if a file path meets the 180 character limit and other rules.
func (v *DirectoryValidator) ValidatePath(path string) []domain.ValidationIssue {
	var issues []domain.ValidationIssue
	
	// Check 180 character limit (2.3.12)
	if len(path) > 180 {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			-1, // directory-level
			"2.3.12",
			fmt.Sprintf("Path exceeds 180 characters (%d)", len(path)),
		))
	}
	
	// Check for leading spaces (2.3.20)
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if len(part) > 0 && part[0] == ' ' {
			issues = append(issues, domain.NewIssue(
				domain.LevelError,
				-1,
				"2.3.20",
				fmt.Sprintf("Leading space not allowed in path component: %q", part),
			))
			break
		}
	}
	
	return issues
}

// ValidateStructure checks directory organization (single disc vs multi-disc).
func (v *DirectoryValidator) ValidateStructure(basePath string, files []string) []domain.ValidationIssue {
	var issues []domain.ValidationIssue
	
	// Analyze directory structure
	hasSubdirs := false
	nestedLevels := 0
	discDirs := make(map[string]bool)
	
	for _, file := range files {
		relPath := strings.TrimPrefix(file, basePath)
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		
		parts := strings.Split(relPath, string(filepath.Separator))
		depth := len(parts) - 1 // subtract 1 for the filename itself
		
		if depth > nestedLevels {
			nestedLevels = depth
		}
		
		if depth > 0 {
			hasSubdirs = true
			// Check if it looks like a disc directory (CD1, CD2, Disc 1, etc.)
			dirName := parts[0]
			if isDiscDirectory(dirName) {
				discDirs[dirName] = true
			}
		}
	}
	
	// Check for unnecessary nesting (2.3.3)
	if nestedLevels > 1 {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			-1,
			"2.3.3",
			fmt.Sprintf("Unnecessary nested folders detected (%d levels deep)", nestedLevels),
		))
	}
	
	// For multi-disc, should have disc subdirectories
	if len(discDirs) > 1 {
		// Valid multi-disc structure
	} else if hasSubdirs && len(discDirs) == 0 {
		// Has subdirectories but they don't look like disc folders
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			-1,
			"2.3.3",
			"Subdirectories detected but not in standard disc format (CD1, CD2, etc.)",
		))
	}
	
	return issues
}

// ValidateFolderName checks if the album folder name follows conventions.
func (v *DirectoryValidator) ValidateFolderName(folderName string, album *domain.Album) []domain.ValidationIssue {
	var issues []domain.ValidationIssue
	
	// Check 180 character limit
	if len(folderName) > 180 {
		issues = append(issues, domain.NewIssue(
			domain.LevelError,
			-1,
			"2.3.12",
			fmt.Sprintf("Folder name exceeds 180 characters (%d)", len(folderName)),
		))
	}
	
	// Check if folder name is meaningful (2.3.2)
	// Minimum is "Album" title, but preferred is "Artist - Album (Year) - Format"
	lowerFolder := strings.ToLower(folderName)
	lowerAlbum := strings.ToLower(album.Title())
	
	if !strings.Contains(lowerFolder, lowerAlbum) {
		issues = append(issues, domain.NewIssue(
			domain.LevelWarning,
			-1,
			"2.3.2",
			"Folder name should contain the album title",
		))
	}
	
	// For classical music, should mention composer (classical.folder_name)
	tracks := album.Tracks()
	if len(tracks) > 0 {
		composer := tracks[0].Composer()
		if composer.Name() != "" {
			composerLastName := lastName(composer.Name())
			if !strings.Contains(lowerFolder, strings.ToLower(composerLastName)) {
				issues = append(issues, domain.NewIssue(
					domain.LevelWarning,
					-1,
					"classical.folder_name",
					fmt.Sprintf("Folder name should mention composer (%s)", composer.Name()),
				))
			}
		}
	}
	
	return issues
}

// isDiscDirectory checks if a directory name looks like a disc folder.
func isDiscDirectory(name string) bool {
	lower := strings.ToLower(name)
	patterns := []string{"cd", "disc", "disk"}
	
	for _, pattern := range patterns {
		if strings.HasPrefix(lower, pattern) {
			// Check if followed by a number
			rest := strings.TrimPrefix(lower, pattern)
			rest = strings.TrimSpace(rest)
			if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
				return true
			}
		}
	}
	
	return false
}

// lastName extracts the last name from a full name.
func lastName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
