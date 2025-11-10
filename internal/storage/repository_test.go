package storage

import (
	"encoding/json"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRepository_SaveAndLoad(t *testing.T) {
	repo := NewRepository()

	// Create torrent directly (no constructors needed)
	torrent := &domain.Torrent{
		RootPath:     "test-album",
		Title:        "Test Album",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Test Label",
			CatalogNumber: "TL123",
			Year:          2013,
		},
		Files: []domain.FileLike{
			&domain.Track{
				File: domain.File{
					Path: "01 - Ave Maria.flac",
				},
				Disc:  1,
				Track: 1,
				Title: "Ave Maria",
				Artists: []domain.Artist{
					{Name: "Anton Bruckner", Role: domain.RoleComposer},
				},
			},
		},
	}

	// Save
	data, err := repo.SaveToJSON(torrent)
	if err != nil {
		t.Fatalf("SaveToJSON error: %v", err)
	}

	// Load
	loaded, err := repo.LoadFromJSON(data)
	if err != nil {
		t.Fatalf("LoadFromJSON error: %v", err)
	}

	// Verify (direct field access)
	if loaded.Title != torrent.Title {
		t.Errorf("Title = %v, want %v", loaded.Title, torrent.Title)
	}
	if loaded.OriginalYear != torrent.OriginalYear {
		t.Errorf("OriginalYear = %v, want %v", loaded.OriginalYear, torrent.OriginalYear)
	}
	if len(loaded.Tracks()) != len(torrent.Tracks()) {
		t.Errorf("Track count = %d, want %d", len(loaded.Tracks()), len(torrent.Tracks()))
	}
}

func TestRepository_JSONFormat(t *testing.T) {
	repo := NewRepository()

	torrent := &domain.Torrent{
		RootPath:     "simple-album",
		Title:        "Simple Album",
		OriginalYear: 2013,
		Files: []domain.FileLike{
			&domain.Track{
				File: domain.File{
					Path: "01 - Work.flac",
				},
				Disc:  1,
				Track: 1,
				Title: "Work",
				Artists: []domain.Artist{
					{Name: "Bach", Role: domain.RoleComposer},
				},
			},
		},
	}

	data, err := repo.SaveToJSON(torrent)
	if err != nil {
		t.Fatalf("SaveToJSON error: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify expected fields exist
	if _, ok := decoded["title"]; !ok {
		t.Error("JSON missing 'title' field")
	}
	if _, ok := decoded["original_year"]; !ok {
		t.Error("JSON missing 'original_year' field")
	}
	if _, ok := decoded["files"]; !ok {
		t.Error("JSON missing 'files' field")
	}
}

func TestRepository_RoleJSON(t *testing.T) {
	// Test that Role enum serializes correctly
	repo := NewRepository()

	torrent := &domain.Torrent{
		RootPath:     "role-test",
		Title:        "Role Test",
		OriginalYear: 2013,
		Files: []domain.FileLike{
			&domain.Track{
				File: domain.File{
					Path: "01 - Test.flac",
				},
				Disc:  1,
				Track: 1,
				Title: "Test",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
					{Name: "Soloist", Role: domain.RoleSoloist},
					{Name: "Ensemble", Role: domain.RoleEnsemble},
					{Name: "Conductor", Role: domain.RoleConductor},
				},
			},
		},
	}

	// Marshal
	data, err := repo.SaveToJSON(torrent)
	if err != nil {
		t.Fatalf("SaveToJSON error: %v", err)
	}

	// Unmarshal
	loaded, err := repo.LoadFromJSON(data)
	if err != nil {
		t.Fatalf("LoadFromJSON error: %v", err)
	}

	// Verify roles round-trip correctly
	tracks := loaded.Tracks()
	if len(tracks) != 1 {
		t.Fatal("Expected 1 track")
	}
	artists := tracks[0].Artists
	if len(artists) != 4 {
		t.Fatalf("Expected 4 artists, got %d", len(artists))
	}

	expectedRoles := []domain.Role{
		domain.RoleComposer,
		domain.RoleSoloist,
		domain.RoleEnsemble,
		domain.RoleConductor,
	}

	for i, expected := range expectedRoles {
		if artists[i].Role != expected {
			t.Errorf("Artist %d role = %v, want %v", i, artists[i].Role, expected)
		}
	}
}
