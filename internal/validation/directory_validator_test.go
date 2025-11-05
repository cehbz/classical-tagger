package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestDirectoryValidator_ValidatePath(t *testing.T) {
	validator := NewDirectoryValidator()

	tests := []struct {
		Name           string
		Path           string
		WantErrorCount int
	}{
		{
			Name:           "valid short path",
			Path:           "/music/Bach - Goldberg Variations (1981) - FLAC/01 Aria.flac",
			WantErrorCount: 0,
		},
		{
			Name:           "path too long",
			Path:           "/" + string(make([]byte, 181)),
			WantErrorCount: 1,
		},
		{
			Name:           "leading space in filename",
			Path:           "/music/album/ 01 Track.flac",
			WantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := validator.ValidatePath(tt.Path)

			errorCount := 0
			for _, issue := range issues {
				if issue.Level == domain.LevelError {
					errorCount++
				}
			}

			if errorCount != tt.WantErrorCount {
				t.Errorf("ValidatePath() error count = %d, want %d", errorCount, tt.WantErrorCount)
				for _, issue := range issues {
					t.Logf("  %s", issue)
				}
			}
		})
	}
}

func TestDirectoryValidator_ValidateStructure(t *testing.T) {
	validator := NewDirectoryValidator()

	tests := []struct {
		Name           string
		Files          []string
		IsMultiDisc    bool
		WantErrorCount int
	}{
		{
			Name: "valid single disc",
			Files: []string{
				"01 Track One.flac",
				"02 Track Two.flac",
			},
			IsMultiDisc:    false,
			WantErrorCount: 0,
		},
		{
			Name: "valid multi disc",
			Files: []string{
				"CD1/01 Track One.flac",
				"CD1/02 Track Two.flac",
				"CD2/01 Track One.flac",
			},
			IsMultiDisc:    true,
			WantErrorCount: 0,
		},
		{
			Name: "invalid nested structure",
			Files: []string{
				"Album/CD1/01 Track.flac",
			},
			IsMultiDisc:    true,
			WantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := validator.ValidateStructure("/base/album", tt.Files)

			errorCount := 0
			for _, issue := range issues {
				if issue.Level == domain.LevelError {
					errorCount++
				}
			}

			if errorCount != tt.WantErrorCount {
				t.Errorf("ValidateStructure() error count = %d, want %d", errorCount, tt.WantErrorCount)
			}
		})
	}
}

func TestDirectoryValidator_ValidateFolderName(t *testing.T) {
	validator := NewDirectoryValidator()

	tests := []struct {
		Name        string
		FolderName  string
		Torrent *domain.Torrent
		WantWarning bool
	}{
		{
			Name:       "good folder name with artist and album",
			FolderName: "Glenn Gould - Goldberg Variations (1981) - FLAC",
			Torrent: &domain.Torrent{
				Title:        "Goldberg Variations",
				OriginalYear: 1981,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Aria",
						Artists: []domain.Artist{
							{Name: "Johann Sebastian Bach", Role: domain.RoleComposer},
						},
					},
				},
			},
			WantWarning: true, // Still warns to mention composer
		},
		{
			Name:       "minimal folder name (just album)",
			FolderName: "Goldberg Variations",
			Torrent: &domain.Torrent{
				Title:        "Goldberg Variations",
				OriginalYear: 1981,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Aria",
						Artists: []domain.Artist{
							{Name: "Bach", Role: domain.RoleComposer},
						},
					},
				},
			},
			WantWarning: true, // Minimal is acceptable per rules but warns about composer
		},
		{
			Name:       "folder name missing composer",
			FolderName: "Piano Works",
			Torrent: &domain.Torrent{
				Title:        "Piano Works",
				OriginalYear: 1981,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Nocturne",
						Artists: []domain.Artist{
							{Name: "Frederic Chopin", Role: domain.RoleComposer},
						},
					},
				},
			},
			WantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := validator.ValidateFolderName(tt.FolderName, tt.Torrent)

			hasWarning := false
			for _, issue := range issues {
				if issue.Level == domain.LevelWarning {
					hasWarning = true
				}
			}

			if hasWarning != tt.WantWarning {
				t.Errorf("ValidateFolderName() warning = %v, want %v", hasWarning, tt.WantWarning)
			}
		})
	}
}

