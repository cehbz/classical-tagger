package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestAlbum_BasicConstruction shows the new mutable construction pattern
func TestAlbum_BasicConstruction(t *testing.T) {
	// Old way (removed):
	// album := domain.Album{Title: "Test Album", OriginalYear: 2013}
	// album.AddTrack(track)

	// New way - direct struct construction:
	album := &domain.Album{
		Title:        "Noël ! Weihnachten ! Christmas!",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Harmonia Mundi",
			CatalogNumber: "HMC902170",
			Year:          2013,
		},
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Frohlocket, ihr Völker auf Erden, Op. 79/1",
				Name:  "01 Frohlocket, ihr Völker auf Erden, Op. 79-1.flac",
				Artists: []domain.Artist{
					{Name: "Felix Mendelssohn Bartholdy", Role: domain.RoleComposer},
					{Name: "RIAS Kammerchor Berlin", Role: domain.RoleEnsemble},
					{Name: "Hans-Christoph Rademann", Role: domain.RoleConductor},
				},
			},
		},
	}

	// Direct field access (no getters)
	if album.Title != "Noël ! Weihnachten ! Christmas!" {
		t.Errorf("Title = %v, want %v", album.Title, "Noël ! Weihnachten ! Christmas!")
	}
	if album.OriginalYear != 2013 {
		t.Errorf("OriginalYear = %v, want 2013", album.OriginalYear)
	}
	if album.Edition == nil {
		t.Fatal("Edition should not be nil")
	}
	if len(album.Tracks) != 1 {
		t.Errorf("Track count = %d, want 1", len(album.Tracks))
	}
}

// TestAlbum_Mutation shows that objects are fully mutable
func TestAlbum_Mutation(t *testing.T) {
	album := &domain.Album{
		Title:        "Original Title",
		OriginalYear: 2013,
		Tracks:       []*domain.Track{},
	}

	// Can mutate directly
	album.Title = "Changed Title"
	album.OriginalYear = 2014

	// Can add tracks directly
	track := &domain.Track{
		Disc:  1,
		Track: 1,
		Title: "Work",
		Artists: []domain.Artist{
			{Name: "Bach", Role: domain.RoleComposer},
		},
	}
	album.Tracks = append(album.Tracks, track)

	if album.Title != "Changed Title" {
		t.Errorf("Title = %v, want 'Changed Title'", album.Title)
	}
	if len(album.Tracks) != 1 {
		t.Errorf("Track count = %d, want 1", len(album.Tracks))
	}
}

// TestAlbum_JSONRoundTrip shows that JSON serialization works directly
func TestAlbum_JSONRoundTrip(t *testing.T) {
	// Create album
	original := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Test Label",
			CatalogNumber: "TL123",
			Year:          2013,
		},
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Symphony No. 1",
				Artists: []domain.Artist{
					{Name: "Brahms", Role: domain.RoleComposer},
					{Name: "Berlin Phil", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	// Marshal to JSON (no DTO needed!)
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal back
	var decoded domain.Album
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify
	if decoded.Title != original.Title {
		t.Errorf("Title = %v, want %v", decoded.Title, original.Title)
	}
	if decoded.OriginalYear != original.OriginalYear {
		t.Errorf("OriginalYear = %v, want %v", decoded.OriginalYear, original.OriginalYear)
	}
	if len(decoded.Tracks) != len(original.Tracks) {
		t.Errorf("Track count = %d, want %d", len(decoded.Tracks), len(original.Tracks))
	}
}

// TestAlbum_Validate shows validation still works
func TestAlbum_Validate(t *testing.T) {
	// Empty album (no tracks) - should have error
	empty := &domain.Album{
		Title:        "Empty Album",
		OriginalYear: 2013,
		Tracks:       []*domain.Track{},
	}

	issues := empty.Validate()
	foundError := false
	for _, issue := range issues {
		if issue.Level == domain.LevelError {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("Expected validation error for empty album")
	}

	// Valid album
	valid := &domain.Album{
		Title:        "Valid Album",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Label",
			CatalogNumber: "CAT123",
			Year:          2013,
		},
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Work",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
					{Name: "Ensemble", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	issues = valid.Validate()
	errorCount := 0
	for _, issue := range issues {
		if issue.Level == domain.LevelError {
			errorCount++
		}
	}
	if errorCount > 0 {
		t.Errorf("Expected no errors for valid album, got %d", errorCount)
	}
}

// TestTrack_Composer shows the Composer() helper still works
func TestTrack_Composer(t *testing.T) {
	track := &domain.Track{
		Disc:  1,
		Track: 1,
		Title: "Symphony",
		Artists: []domain.Artist{
			{Name: "Beethoven", Role: domain.RoleComposer},
			{Name: "Berlin Phil", Role: domain.RoleEnsemble},
			{Name: "Karajan", Role: domain.RoleConductor},
		},
	}

	composers := track.Composers()
	if len(composers) != 1 {
		t.Errorf("Expected 1 composer, got %d", len(composers))
	}
	composer := composers[0]
	if composer.Name != "Beethoven" {
		t.Errorf("Composer name = %v, want 'Beethoven'", composer.Name)
	}
}
