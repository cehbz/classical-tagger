package domain

import (
	"strings"
	"testing"
)

func TestNewAlbum(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		originalYear int
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid album",
			title:        "Noël ! Weihnachten ! Christmas!",
			originalYear: 2013,
			wantErr:      false,
		},
		{
			name:         "empty title",
			title:        "",
			originalYear: 2013,
			wantErr:      false,
		},
		{
			name:         "whitespace title",
			title:        "   ",
			originalYear: 2013,
			wantErr:      false,
		},
		{
			name:         "zero year allowed",
			title:        "Some Album",
			originalYear: 0,
			wantErr:      false,
		},
		{
			name:         "negative year",
			title:        "Some Album",
			originalYear: -1,
			wantErr:      true,
			errMsg:       "album year must be >= 0, got -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAlbum(tt.title, tt.originalYear)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAlbum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewAlbum() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				if got.Title() != tt.title {
					t.Errorf("NewAlbum().Title() = %v, want %v", got.Title(), tt.title)
				}
				if got.OriginalYear() != tt.originalYear {
					t.Errorf("NewAlbum().OriginalYear() = %v, want %v", got.OriginalYear(), tt.originalYear)
				}
				if got.Edition() != nil {
					t.Errorf("NewAlbum().Edition() = %v, want nil", got.Edition())
				}
			}
		})
	}
}

func TestAlbum_WithEdition(t *testing.T) {
	album, _ := NewAlbum("Test Album", 2013)
	edition, _ := NewEdition("test edition", 2013)
	edition = edition.WithCatalogNumber("TE902170")

	album = album.WithEdition(edition)

	if album.Edition() == nil {
		t.Fatal("Album.Edition() should not be nil after WithEdition")
	}
	if album.Edition().Label() != "test edition" {
		t.Errorf("Album.Edition().Label() = %v, want %v", album.Edition().Label(), "test edition")
	}
}

func TestAlbum_AddTrack(t *testing.T) {
	album, _ := NewAlbum("Test Album", 2013)
	composer, _ := NewArtist("Felix Mendelssohn Bartholdy", RoleComposer)
	track, _ := NewTrack(1, 1, "Test Work", []Artist{composer})

	err := album.AddTrack(track)
	if err != nil {
		t.Errorf("Album.AddTrack() unexpected error = %v", err)
	}

	tracks := album.Tracks()
	if len(tracks) != 1 {
		t.Errorf("Album.Tracks() length = %d, want 1", len(tracks))
	}
}

func TestAlbum_Validate_NoTracks(t *testing.T) {
	album, _ := NewAlbum("Empty Album", 2013)

	issues := album.Validate()

	foundError := false
	for _, issue := range issues {
		if strings.Contains(issue.Message(), "at least one track") {
			foundError = true
			if issue.Level() != LevelError {
				t.Errorf("No tracks error should be ERROR level, got %v", issue.Level())
			}
		}
	}

	if !foundError {
		t.Error("Expected validation error for album with no tracks")
	}
}

func TestAlbum_Validate_MissingEdition(t *testing.T) {
	album, _ := NewAlbum("Test Album", 2013)
	composer, _ := NewArtist("Johannes Brahms", RoleComposer)
	track, _ := NewTrack(1, 1, "Symphony No. 1", []Artist{composer})
	album.AddTrack(track)

	issues := album.Validate()

	foundWarning := false
	for _, issue := range issues {
		if strings.Contains(issue.Message(), "Edition information") {
			foundWarning = true
			if issue.Level() != LevelWarning {
				t.Errorf("Missing edition should be WARNING level, got %v", issue.Level())
			}
		}
	}

	if !foundWarning {
		t.Error("Expected validation warning for missing edition")
	}
}

func TestAlbum_Validate_AllIssues(t *testing.T) {
	// Create album with tracks that have validation issues
	album, _ := NewAlbum("Test Album", 2013)
	composer, _ := NewArtist("Johann Sebastian Bach", RoleComposer)

	// Track with composer in title (should generate error)
	track1, _ := NewTrack(1, 1, "Bach: Goldberg Variations", []Artist{composer})
	album.AddTrack(track1)

	// Valid track
	track2, _ := NewTrack(1, 2, "Goldberg Variations, BWV 988", []Artist{composer})
	album.AddTrack(track2)

	issues := album.Validate()

	// Should have:
	// - 1 warning for missing edition
	// - 1 error from track1 (composer in title)
	if len(issues) < 2 {
		t.Errorf("Album.Validate() returned %d issues, expected at least 2", len(issues))
	}

	// Check that issues from tracks are included
	hasTrackError := false
	for _, issue := range issues {
		if issue.Track() > 0 && issue.Level() == LevelError {
			hasTrackError = true
		}
	}

	if !hasTrackError {
		t.Error("Expected track-level error to be included in album validation")
	}
}
