package filesystem

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestDirectoryValidator_ValidatePath(t *testing.T) {
	validator := NewDirectoryValidator()
	
	tests := []struct {
		name           string
		path           string
		wantErrorCount int
	}{
		{
			name:           "valid short path",
			path:           "/music/Bach - Goldberg Variations (1981) - FLAC/01 Aria.flac",
			wantErrorCount: 0,
		},
		{
			name:           "path too long",
			path:           "/" + string(make([]byte, 181)),
			wantErrorCount: 1,
		},
		{
			name:           "leading space in filename",
			path:           "/music/album/ 01 Track.flac",
			wantErrorCount: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.ValidatePath(tt.path)
			
			errorCount := 0
			for _, issue := range issues {
				if issue.Level() == domain.LevelError {
					errorCount++
				}
			}
			
			if errorCount != tt.wantErrorCount {
				t.Errorf("ValidatePath() error count = %d, want %d", errorCount, tt.wantErrorCount)
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
		name           string
		files          []string
		isMultiDisc    bool
		wantErrorCount int
	}{
		{
			name: "valid single disc",
			files: []string{
				"01 Track One.flac",
				"02 Track Two.flac",
			},
			isMultiDisc:    false,
			wantErrorCount: 0,
		},
		{
			name: "valid multi disc",
			files: []string{
				"CD1/01 Track One.flac",
				"CD1/02 Track Two.flac",
				"CD2/01 Track One.flac",
			},
			isMultiDisc:    true,
			wantErrorCount: 0,
		},
		{
			name: "invalid nested structure",
			files: []string{
				"Album/CD1/01 Track.flac",
			},
			isMultiDisc:    true,
			wantErrorCount: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.ValidateStructure("/base/album", tt.files)
			
			errorCount := 0
			for _, issue := range issues {
				if issue.Level() == domain.LevelError {
					errorCount++
				}
			}
			
			if errorCount != tt.wantErrorCount {
				t.Errorf("ValidateStructure() error count = %d, want %d", errorCount, tt.wantErrorCount)
			}
		})
	}
}

func TestDirectoryValidator_ValidateFolderName(t *testing.T) {
	validator := NewDirectoryValidator()
	
	tests := []struct {
		name        string
		folderName  string
		album       *domain.Album
		wantWarning bool
	}{
		{
			name:       "good folder name with artist and album",
			folderName: "Glenn Gould - Goldberg Variations (1981) - FLAC",
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Goldberg Variations", 1981)
				composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer})
				album.AddTrack(track)
				return album
			}(),
			wantWarning: false,
		},
		{
			name:       "minimal folder name (just album)",
			folderName: "Goldberg Variations",
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Goldberg Variations", 1981)
				composer, _ := domain.NewArtist("Bach", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer})
				album.AddTrack(track)
				return album
			}(),
			wantWarning: false, // Minimal is acceptable per rules
		},
		{
			name:       "folder name missing composer",
			folderName: "Piano Works",
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Piano Works", 1981)
				composer, _ := domain.NewArtist("Frederic Chopin", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Nocturne", []domain.Artist{composer})
				album.AddTrack(track)
				return album
			}(),
			wantWarning: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.ValidateFolderName(tt.folderName, tt.album)
			
			hasWarning := false
			for _, issue := range issues {
				if issue.Level() == domain.LevelWarning {
					hasWarning = true
				}
			}
			
			if hasWarning != tt.wantWarning {
				t.Errorf("ValidateFolderName() warning = %v, want %v", hasWarning, tt.wantWarning)
			}
		})
	}
}
