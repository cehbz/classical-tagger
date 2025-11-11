package discogs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestClient_Search(t *testing.T) {
	// Mock Discogs API response
	mockResponse := `{
		"results": [
			{
				"id": 195873,
				"title": "Goldberg Variations",
				"year": "2013",
				"label": ["Deutsche Grammophon"],
				"catno": "479 1234",
				"format": ["CD", "Album"],
				"country": "Germany"
			},
			{
				"id": 842951,
				"title": "Goldberg-Variationen",  
				"year": "1985",
				"label": ["Archiv Produktion"],
				"catno": "415 130-2",
				"format": ["CD", "Album"],
				"country": "Germany"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/database/search" {
			t.Errorf("Expected path /database/search, got %s", r.URL.Path)
		}

		// Check auth header
		auth := r.Header.Get("Authorization")
		if auth != "Discogs token=test-token" {
			t.Errorf("Expected auth header, got %s", auth)
		}

		// Check query params
		q := r.URL.Query()
		if q.Get("artist") != "Bach" {
			t.Errorf("Expected artist=Bach, got %s", q.Get("artist"))
		}
		if q.Get("release_title") != "Goldberg Variations" {
			t.Errorf("Expected release_title=Goldberg Variations, got %s", q.Get("release_title"))
		}
		if q.Get("type") != "release" {
			t.Errorf("Expected type=release, got %s", q.Get("type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	releases, err := client.Search("Bach", "Goldberg Variations")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("Expected 2 releases, got %d", len(releases))
	}

	// Verify first release
	if releases[0].ID != 195873 {
		t.Errorf("Expected ID 195873, got %d", releases[0].ID)
	}
	if releases[0].Title != "Goldberg Variations" {
		t.Errorf("Expected title 'Goldberg Variations', got %s", releases[0].Title)
	}
	if releases[0].Label != "Deutsche Grammophon" {
		t.Errorf("Expected label 'Deutsche Grammophon', got %s", releases[0].Label)
	}
}

func TestClient_Search_NoResults(t *testing.T) {
	mockResponse := `{"results": []}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	releases, err := client.Search("Unknown Artist", "Unknown Album")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(releases) != 0 {
		t.Errorf("Expected 0 releases, got %d", len(releases))
	}
}

func TestClient_GetRelease(t *testing.T) {
	// Mock detailed release response
	mockResponse := `{
		"id": 195873,
		"title": "Goldberg Variations",
		"artists": [
			{"name": "Johann Sebastian Bach", "role": "Composer"},
			{"name": "Glenn Gould", "role": "Piano"}
		],
		"year": 2013,
		"labels": [
			{"name": "Deutsche Grammophon", "catno": "479 1234"}
		],
		"tracklist": [
			{
				"position": "1",
				"title": "Aria",
				"duration": "4:35"
			},
			{
				"position": "2", 
				"title": "Variation 1",
				"duration": "1:55"
			}
		],
		"extraartists": [
			{"name": "Producer Name", "role": "Producer"},
			{"name": "Engineer Name", "role": "Engineer"}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/195873" {
			t.Errorf("Expected path /releases/195873, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	release, err := client.GetRelease(195873)
	if err != nil {
		t.Fatalf("GetRelease() error = %v", err)
	}

	if release.ID != 195873 {
		t.Errorf("Expected ID 195873, got %d", release.ID)
	}
	if release.Title != "Goldberg Variations" {
		t.Errorf("Expected title 'Goldberg Variations', got %s", release.Title)
	}
	if len(release.Artists) != 2 {
		t.Errorf("Expected 2 artists, got %d", len(release.Artists))
	}
	if len(release.Tracklist) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(release.Tracklist))
	}
}

func TestClient_GetRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Release not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	release, err := client.GetRelease(999999)
	if err == nil {
		t.Error("Expected error for not found release")
	}
	if release != nil {
		t.Error("Expected nil release for not found")
	}
}

func TestRelease_MarshalJSON(t *testing.T) {
	release := &Release{
		ID:            123,
		Title:         "Test Album",
		Year:          2020,
		Label:         "Test Label",
		CatalogNumber: "CAT-123",
		Artists: []Artist{
			{Name: "Composer", Role: "Composer"},
		},
		Tracklist: []Track{
			{Position: "1", Title: "Track 1", Duration: "3:45"},
		},
	}

	data, err := json.Marshal(release)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Release
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != release.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, release.ID)
	}
	if decoded.Title != release.Title {
		t.Errorf("Title mismatch: got %s, want %s", decoded.Title, release.Title)
	}
}

func TestRelease_DomainTorrent(t *testing.T) {
	tests := []struct {
		name    string
		release *Release
		want    *domain.Torrent
	}{
		{
			name: "smoke test",
			release: &Release{
				ID:            123,
				Title:         "Test Album",
				Year:          2020,
				Label:         "Test Label",
				CatalogNumber: "CAT-123",
				Artists:       []Artist{{Name: "Composer", Role: "Composer"}},
				Tracklist:     []Track{{Position: "1", Title: "Track 1", Duration: "3:45"}},
			},
			want: &domain.Torrent{
				RootPath:     "test-path/Composer - Test Album - 2020 [FLAC]",
				Title:        "Test Album",
				OriginalYear: 2020,
				Edition:      &domain.Edition{Label: "Test Label", CatalogNumber: "CAT-123", Year: 2020},
				AlbumArtist:  []domain.Artist{{Name: "Composer", Role: domain.RoleComposer}},
				Files:        []domain.FileLike{&domain.Track{Disc: 1, Track: 1, Title: "Track 1"}},
			},
		},
		{
			name: "test with subtracks",
			release: &Release{
				ID:            123,
				Title:         "Test Album",
				Year:          2020,
				Label:         "Test Label",
				CatalogNumber: "CAT-123",
				Artists:       []Artist{{Name: "Ludwig von Beethoven", Role: "Composer"}},
				Tracklist: []Track{
					{Position: "1", Title: "Track 1", Duration: "3:45"},
					{Position: "", Title: "Parent Track",
						SubTracks: []Track{{Position: "2", Title: "Subtrack 1", Artists: []Artist{{Name: "E Power Biggs", Role: "Soloist"}, {Name: "J. S. Bach", Role: "Composer"}}}}},
				},
			},
			want: &domain.Torrent{
				RootPath:     "test-path/Beethoven - Test Album (Biggs) - 2020 [FLAC]",
				Title:        "Test Album",
				OriginalYear: 2020,
				Edition:      &domain.Edition{Label: "Test Label", CatalogNumber: "CAT-123", Year: 2020},
				AlbumArtist:  []domain.Artist{{Name: "Ludwig von Beethoven", Role: domain.RoleComposer}},
				Files: []domain.FileLike{
					&domain.Track{Disc: 1, Track: 1, Title: "Track 1"},
					&domain.Track{Disc: 1, Track: 2, Title: "Parent Track: Subtrack 1",
						Artists: []domain.Artist{{Name: "E Power Biggs", Role: domain.RoleSoloist}, {Name: "J. S. Bach", Role: domain.RoleComposer}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			torrent := tt.release.DomainTorrent("test-path")
			if torrent == nil {
				t.Fatal("DomainTorrent returned nil")
			}
			// deep compare torrent and want, ignoring zero values in want
			if tt.want.RootPath != "" && tt.want.RootPath != torrent.RootPath {
				t.Errorf("RootPath mismatch: got %s, want %s", torrent.RootPath, tt.want.RootPath)
			}
			if tt.want.Title != "" && tt.want.Title != torrent.Title {
				t.Errorf("Title mismatch: got %s, want %s", torrent.Title, tt.want.Title)
			}
			if tt.want.OriginalYear != 0 && tt.want.OriginalYear != torrent.OriginalYear {
				t.Errorf("OriginalYear mismatch: got %d, want %d", torrent.OriginalYear, tt.want.OriginalYear)
			}
			if tt.want.Edition != nil {
				if tt.want.Edition.Label != "" && tt.want.Edition.Label != torrent.Edition.Label {
					t.Errorf("Edition.Label mismatch: got %s, want %s", torrent.Edition.Label, tt.want.Edition.Label)
				}
				if tt.want.Edition.CatalogNumber != "" && tt.want.Edition.CatalogNumber != torrent.Edition.CatalogNumber {
					t.Errorf("Edition.CatalogNumber mismatch: got %s, want %s", torrent.Edition.CatalogNumber, tt.want.Edition.CatalogNumber)
				}
				if tt.want.Edition.Year != 0 && tt.want.Edition.Year != torrent.Edition.Year {
					t.Errorf("Edition.Year mismatch: got %d, want %d", torrent.Edition.Year, tt.want.Edition.Year)
				}
			}
			if tt.want.AlbumArtist != nil {
				wantArtists := make(map[domain.Artist]struct{})
				for _, artist := range tt.want.AlbumArtist {
					wantArtists[artist] = struct{}{}
				}
				gotArtists := make(map[domain.Artist]struct{})
				for _, artist := range torrent.AlbumArtist {
					gotArtists[artist] = struct{}{}
				}
				for artist := range gotArtists {
					if _, ok := wantArtists[artist]; !ok {
						t.Errorf("AlbumArtist mismatch: got %+v but not wanted", artist)
					}
				}
				for artist := range wantArtists {
					if _, ok := gotArtists[artist]; !ok {
						t.Errorf("AlbumArtist mismatch: wanted %+v but didn't get", artist)
					}
				}
			}
			if tt.want.Files != nil && len(tt.want.Files) != len(torrent.Files) {
				t.Errorf("Files mismatch: got %+v, want %+v", len(torrent.Files), len(tt.want.Files))
			}
			if tt.want.SiteMetadata != nil && tt.want.SiteMetadata != torrent.SiteMetadata {
				t.Errorf("SiteMetadata mismatch: got %+v, want %+v", torrent.SiteMetadata, tt.want.SiteMetadata)
			}
		})
	}
}
