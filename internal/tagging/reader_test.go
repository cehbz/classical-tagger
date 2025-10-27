package tagging

import (
	"errors"
	"path/filepath"
	"testing"
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

	// Commented out until we have test fixtures
	// reader := NewFLACReader()
	// track, err := reader.ReadTrackFromFile("testdata/01-test.flac", 1, 1)
	//
	// if err != nil {
	// 	t.Fatalf("ReadTrackFromFile() error = %v", err)
	// }
	//
	// if track.Title == "" {
	// 	t.Error("Expected non-empty title")
	// }
}

func TestValidateExpectedNumbers(t *testing.T) {
	// Build a fake domain.Track via Metadata -> ToTrack to avoid direct construction
	md := Metadata{
		Title:       "Track Title",
		Composer:    "Johann Sebastian Bach",
		Artist:      "Performer",
		Album:       "Album",
		Year:        "1981",
		TrackNumber: "3",
		DiscNumber:  "2",
	}
	tr, err := md.ToTrack("03 Example.flac")
	if err != nil {
		t.Fatalf("setup ToTrack failed: %v", err)
	}

	t.Run("happy path - matches expected disc and track", func(t *testing.T) {
		if err := validateDiskAndTrackNumbers(tr, 2, 3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("mismatch track - returns error", func(t *testing.T) {
		err := validateDiskAndTrackNumbers(tr, 2, 4)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, err) { // sanity to use err
			// no-op
		}
	})

	t.Run("mismatch disc - returns error", func(t *testing.T) {
		if err := validateDiskAndTrackNumbers(tr, 1, 3); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

// TestMetadata validates the Metadata structure
func TestMetadata(t *testing.T) {
	tests := []struct {
		Name     string
		Metadata Metadata
		WantErr  bool
	}{
		{
			Name: "valid metadata",
			Metadata: Metadata{
				Title:       "Goldberg Variations",
				Artist:      "Glenn Gould",
				Album:       "Bach: Goldberg Variations",
				Composer:    "Johann Sebastian Bach",
				Year:        "1981",
				TrackNumber: "1",
			},
			WantErr: false,
		},
		{
			Name: "missing required field",
			Metadata: Metadata{
				Title:  "Some Work",
				Artist: "",
				Album:  "Test Album",
			},
			WantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := tt.Metadata.Validate()
			if (err != nil) != tt.WantErr {
				t.Errorf("Metadata.Validate() error = %v, wantErr %v", err, tt.WantErr)
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

	if track.Title != metadata.Title {
		t.Errorf("ToTrack() title = %v, want %v", track.Title, metadata.Title)
	}

	composers := track.Composers()
	if len(composers) == 0 {
		t.Errorf("ToTrack() no composers found")
	}
	composer := composers[0]
	if composer.Name != metadata.Composer {
		t.Errorf("ToTrack() composer = %v, want %v", composer.Name, metadata.Composer)
	}
}
