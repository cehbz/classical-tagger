package tagging

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestFLACReader_ReadFile(t *testing.T) {
	// Note: These tests require actual FLAC files to run.
	// For now, we test the interface and error handling.
	
	reader := NewFLACReader()
	
	t.Run("non-existent file", func(t *testing.T) {
		_, err := reader.ReadFile("nonexistent.flac")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestFLACReader_ReadTrackFromFile(t *testing.T) {
	// This test would need a real FLAC file with proper tags
	// For CI/CD, you'd want to include a test fixture
	
	t.Skip("Requires FLAC test fixture")
	
	reader := NewFLACReader()
	track, err := reader.ReadTrackFromFile("testdata/01-test.flac", 1, 1)
	
	if err != nil {
		t.Fatalf("ReadTrackFromFile() error = %v", err)
	}
	
	if track.Title() == "" {
		t.Error("Expected non-empty title")
	}
}

// TestMetadata validates the Metadata structure
func TestMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		wantErr  bool
	}{
		{
			name: "valid metadata",
			metadata: Metadata{
				Title:       "Goldberg Variations",
				Artist:      "Glenn Gould",
				Album:       "Bach: Goldberg Variations",
				Composer:    "Johann Sebastian Bach",
				Year:        "1981",
				TrackNumber: "1",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			metadata: Metadata{
				Title:  "Some Work",
				Artist: "",
				Album:  "Test Album",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Metadata.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetadata_ToTrack(t *testing.T) {
	metadata := Metadata{
		Title:       "Symphony No. 5, Op. 67: I. Allegro con brio",
		Composer:    "Ludwig van Beethoven",
		Artist:      "Vienna Philharmonic, Carlos Kleiber",
		Album:       "Beethoven: Symphonies 5 & 7",
		Year:        "1976",
		TrackNumber: "1",
		DiscNumber:  "1",
	}
	
	track, err := metadata.ToTrack(filepath.Base("01 Symphony No. 5.flac"))
	if err != nil {
		t.Fatalf("ToTrack() error = %v", err)
	}
	
	if track.Title() != metadata.Title {
		t.Errorf("ToTrack() title = %v, want %v", track.Title(), metadata.Title)
	}
	
	composer := track.Composer()
	if composer.Name() != metadata.Composer {
		t.Errorf("ToTrack() composer = %v, want %v", composer.Name(), metadata.Composer)
	}
}
