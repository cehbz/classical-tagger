package storage

import (
	"encoding/json"
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestAlbumDTO_ToAlbum(t *testing.T) {
	dto := AlbumDTO{
		Title:        "Noël ! Weihnachten ! Christmas!",
		OriginalYear: 2013,
		Edition: &EditionDTO{
			Label:         "test label",
			CatalogNumber: "HMC902170",
			EditionYear:   2013,
		},
		Tracks: []TrackDTO{
			{
				Disc:  1,
				Track: 1,
				Title: "Frohlocket, ihr Völker auf Erden, Op. 79/1",
				Composer: ArtistDTO{
					Name: "Felix Mendelssohn Bartholdy",
					Role: "composer",
				},
				Artists: []ArtistDTO{
					{
						Name: "RIAS Kammerchor Berlin",
						Role: "ensemble",
					},
					{
						Name: "Hans-Christoph Rademann",
						Role: "conductor",
					},
				},
				Name: "01 Frohlocket, ihr Völker auf Erden, Op. 79-1.flac",
			},
		},
	}
	
	album, err := dto.ToAlbum()
	if err != nil {
		t.Fatalf("ToAlbum() error = %v", err)
	}
	
	if album.Title() != dto.Title {
		t.Errorf("Title = %v, want %v", album.Title(), dto.Title)
	}
	if album.OriginalYear() != dto.OriginalYear {
		t.Errorf("OriginalYear = %v, want %v", album.OriginalYear(), dto.OriginalYear)
	}
	if album.Edition() == nil {
		t.Fatal("Edition should not be nil")
	}
	if len(album.Tracks()) != len(dto.Tracks) {
		t.Errorf("Track count = %d, want %d", len(album.Tracks()), len(dto.Tracks))
	}
}

func TestAlbumDTO_FromAlbum(t *testing.T) {
	// Create domain album
	album, _ := domain.NewAlbum("Test Album", 2013)
	edition, _ := domain.NewEdition("test label", 2013)
	edition = edition.WithCatalogNumber("HMC902170")
	album = album.WithEdition(edition)
	
	composer, _ := domain.NewArtist("Felix Mendelssohn", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("RIAS Kammerchor", domain.RoleEnsemble)
	track, _ := domain.NewTrack(1, 1, "Test Work", []domain.Artist{composer, ensemble})
	track = track.WithName("01 Test Work.flac")
	album.AddTrack(track)
	
	// Convert to DTO
	dto := FromAlbum(album)
	
	if dto.Title != album.Title() {
		t.Errorf("DTO.Title = %v, want %v", dto.Title, album.Title())
	}
	if dto.Edition == nil {
		t.Fatal("DTO.Edition should not be nil")
	}
	if len(dto.Tracks) != len(album.Tracks()) {
		t.Errorf("DTO track count = %d, want %d", len(dto.Tracks), len(album.Tracks()))
	}
	
	// Check composer in track
	if dto.Tracks[0].Composer.Name != "Felix Mendelssohn" {
		t.Errorf("Composer name = %v, want %v", dto.Tracks[0].Composer.Name, "Felix Mendelssohn")
	}
}

func TestJSON_RoundTrip(t *testing.T) {
	// Create album
	album, _ := domain.NewAlbum("Test Album", 2013)
	composer, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Symphony No. 1", []domain.Artist{composer})
	album.AddTrack(track)
	
	// Convert to DTO and marshal
	dto := FromAlbum(album)
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	
	// Unmarshal and convert back
	var dto2 AlbumDTO
	if err := json.Unmarshal(data, &dto2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	
	album2, err := dto2.ToAlbum()
	if err != nil {
		t.Fatalf("ToAlbum error: %v", err)
	}
	
	// Verify
	if album2.Title() != album.Title() {
		t.Errorf("Round-trip title = %v, want %v", album2.Title(), album.Title())
	}
	if len(album2.Tracks()) != len(album.Tracks()) {
		t.Errorf("Round-trip track count = %d, want %d", len(album2.Tracks()), len(album.Tracks()))
	}
}

func TestRepository_SaveAndLoad(t *testing.T) {
	repo := NewRepository()
	
	// Create album
	album, _ := domain.NewAlbum("Test Album", 2013)
	composer, _ := domain.NewArtist("Anton Bruckner", domain.RoleComposer)
	track, _ := domain.NewTrack(1, 1, "Ave Maria", []domain.Artist{composer})
	album.AddTrack(track)
	
	// Save
	data, err := repo.SaveToJSON(album)
	if err != nil {
		t.Fatalf("SaveToJSON error: %v", err)
	}
	
	// Load
	album2, err := repo.LoadFromJSON(data)
	if err != nil {
		t.Fatalf("LoadFromJSON error: %v", err)
	}
	
	// Verify
	if album2.Title() != album.Title() {
		t.Errorf("Loaded title = %v, want %v", album2.Title(), album.Title())
	}
}
