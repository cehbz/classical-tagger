package discogs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
