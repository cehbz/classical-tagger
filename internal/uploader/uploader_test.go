// internal/uploader/uploader_test.go
package uploader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/ratelimit"
)

func TestRedactedClient_GetTorrent(t *testing.T) {
	tests := []struct {
		name         string
		torrentID    int
		response     string
		statusCode   int
		wantErr      bool
		validateFunc func(*testing.T, *Torrent)
	}{
		{
			name:       "successful fetch",
			torrentID:  123456,
			statusCode: http.StatusOK,
			response: `{
				"status": "success",
				"response": {
					"group": {
						"id": 98765,
						"name": "Christmas Album",
						"year": 2013,
						"tags": ["classical", "choral", "sacred"]
					},
					"torrent": {
						"id": 123456,
						"format": "FLAC",
						"encoding": "Lossless",
						"media": "CD",
						"remastered": false,
						"description": "Original upload notes here",
						"fileList": "01-Track.flac{{{123456}}}02-Track.flac{{{234567}}}"
					}
				}
			}`,
			validateFunc: func(t *testing.T, tm *Torrent) {
				if tm.GroupID != 98765 {
					t.Errorf("expected GroupID 98765, got %d", tm.GroupID)
				}
				if tm.Format != "FLAC" {
					t.Errorf("expected format FLAC, got %s", tm.Format)
				}
				if tm.Description != "Original upload notes here" {
					t.Errorf("unexpected description: %s", tm.Description)
				}
				if len(tm.Tags) != 3 {
					t.Errorf("expected 3 tags, got %d", len(tm.Tags))
				}
			},
		},
		{
			name:       "torrent not found",
			torrentID:  999999,
			statusCode: http.StatusNotFound,
			response:   `{"status": "failure", "error": "bad id parameter"}`,
			wantErr:    true,
		},
		{
			name:       "rate limited",
			torrentID:  123456,
			statusCode: http.StatusTooManyRequests,
			response:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify auth header
				auth := r.Header.Get("Authorization")
				if auth == "" {
					t.Error("missing Authorization header")
				}

				// Verify endpoint
				expectedPath := "/ajax.php"
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify query params
				if r.URL.Query().Get("action") != "torrent" {
					t.Errorf("expected action=torrent, got %s", r.URL.Query().Get("action"))
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "5")
				}
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &RedactedClient{
				BaseURL:     server.URL,
				APIKey:      "test-key",
				HTTPClient:  &http.Client{Timeout: 10 * time.Second},
				RateLimiter: ratelimit.NewRateLimiter(10, 10*time.Second),
			}

			result, err := client.GetTorrent(context.Background(), tt.torrentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTorrent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

func TestRedactedClient_GetTorrentGroup(t *testing.T) {
	tests := []struct {
		name         string
		groupID      int
		response     string
		statusCode   int
		wantErr      bool
		validateFunc func(*testing.T, *TorrentGroup)
	}{
		{
			name:       "successful fetch with detailed artists",
			groupID:    98765,
			statusCode: http.StatusOK,
			response: `{
				"status": "success",
				"response": {
					"group": {
						"id": 98765,
						"name": "Christmas Album",
						"year": 2013,
						"musicInfo": {
							"composers": [
								{"id": 10, "name": "Felix Mendelssohn"},
								{"id": 11, "name": "Johannes Brahms"}
							],
							"conductor": [
								{"id": 2, "name": "Hans-Christoph Rademann"}
							],
							"artists": [
								{"id": 1, "name": "RIAS Kammerchor"}
							]
						},
						"tags": ["classical", "choral"],
						"wikiBody": "Full wiki text here"
					},
					"torrents": [
						{
							"id": 123456,
							"format": "FLAC",
							"encoding": "Lossless",
							"media": "CD",
							"remastered": false
						}
					]
				}
			}`,
			validateFunc: func(t *testing.T, gm *TorrentGroup) {
				if len(gm.Composers) != 2 {
					t.Errorf("expected 2 composers, got %d", len(gm.Composers))
				}
				if len(gm.Conductors) != 1 {
					t.Errorf("expected 1 conductor, got %d", len(gm.Conductors))
				}
				if gm.Conductors[0].Name != "Hans-Christoph Rademann" {
					t.Errorf("unexpected conductor: %s", gm.Conductors[0].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &RedactedClient{
				BaseURL:     server.URL,
				APIKey:      "test-key",
				HTTPClient:  &http.Client{Timeout: 10 * time.Second},
				RateLimiter: ratelimit.NewRateLimiter(10, 10*time.Second),
			}

			result, err := client.GetTorrentGroup(context.Background(), tt.groupID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTorrentGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

func TestUploadCommand_ValidateArtists(t *testing.T) {
	tests := []struct {
		name            string
		redactedArtists []ArtistCredit
		taggedArtists   map[domain.Artist]struct{}
		wantErrors      int
	}{
		{
			name: "exact match",
			redactedArtists: []ArtistCredit{
				{Name: "RIAS Kammerchor", Role: "artists"},
				{Name: "Hans-Christoph Rademann", Role: "conductor"},
				{Name: "Felix Mendelssohn", Role: "composer"},
			},
			taggedArtists: map[domain.Artist]struct{}{
				{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble}:          {},
				{Name: "Hans-Christoph Rademann", Role: domain.RoleConductor}: {},
				{Name: "Felix Mendelssohn", Role: domain.RoleComposer}:        {},
			},
			wantErrors: 0,
		},
		{
			name: "role conflict",
			redactedArtists: []ArtistCredit{
				{Name: "Hans-Christoph Rademann", Role: "conductor"},
			},
			taggedArtists: map[domain.Artist]struct{}{
				{Name: "Hans-Christoph Rademann", Role: domain.RoleComposer}: {}, // Wrong role
			},
			wantErrors: 1,
		},
		{
			name: "missing artist in tags",
			redactedArtists: []ArtistCredit{
				{Name: "RIAS Kammerchor", Role: "artists"},
				{Name: "Missing Artist", Role: "conductor"},
			},
			taggedArtists: map[domain.Artist]struct{}{
				{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble}: {},
			},
			wantErrors: 1,
		},
		{
			name: "extra artist in tags",
			redactedArtists: []ArtistCredit{
				{Name: "RIAS Kammerchor", Role: "artists"},
			},
			taggedArtists: map[domain.Artist]struct{}{
				{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble}: {},
				{Name: "Extra Artist", Role: domain.RoleConductor}:   {}, // Not in Redacted - allowed (superset)
			},
			wantErrors: 0, // Extra artists are allowed (superset)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &UploadCommand{}
			errors := cmd.validateArtistsSuperset(tt.redactedArtists, tt.taggedArtists)

			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d: %v", tt.wantErrors, len(errors), errors)
			}
		})
	}
}

func TestUploadCommand_MergeMetadata(t *testing.T) {
	torrentMeta := &Torrent{
		GroupID:     98765,
		Format:      "FLAC",
		Encoding:    "Lossless",
		Media:       "CD",
		Description: "Original description",
		Tags:        []string{"classical", "choral"},
	}

	groupMeta := &TorrentGroup{
		Composers: []ArtistCredit{
			{Name: "Felix Mendelssohn", Role: "composer"},
		},
		Conductors: []ArtistCredit{
			{Name: "Hans-Christoph Rademann", Role: "conductor"},
		},
		Artists: []ArtistCredit{
			{Name: "RIAS Kammerchor", Role: "artists"},
		},
	}

	localTorrent := &domain.Torrent{
		Title:        "Christmas Album",
		OriginalYear: 2013,
		AlbumArtist: []domain.Artist{
			{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble},
		},
		Edition: &domain.Edition{
			Label:         "Harmonia Mundi",
			CatalogNumber: "HMC 902170",
			Year:          2013,
		},
		Files: []domain.FileLike{
			&domain.Track{
				File:  domain.File{Path: "01-Track.flac"},
				Track: 1,
				Title: "First Track",
				Artists: []domain.Artist{
					{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	trumpReason := "Fixed incorrect composer tags"

	cmd := &UploadCommand{}
	result := cmd.mergeMetadata(torrentMeta, groupMeta, localTorrent, trumpReason)

	// Verify description was appended
	expectedDesc := "Original description\n\n[Trump Upload] Fixed: Fixed incorrect composer tags"
	if result.Description != expectedDesc {
		t.Errorf("expected description:\n%s\ngot:\n%s", expectedDesc, result.Description)
	}

	// Verify artists were merged
	if len(result.Artists) != 1 {
		t.Errorf("expected 1 artist, got %d", len(result.Artists))
	}

	// Verify format info preserved
	if result.Format != "FLAC" {
		t.Errorf("expected FLAC format, got %s", result.Format)
	}
}

func TestUploadCommand_CreateTorrentFile(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "01-Track.flac")
	if err := os.WriteFile(testFile, []byte("fake flac data"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &UploadCommand{
		CacheDir: t.TempDir(),
	}

	torrentPath, err := cmd.createTorrentFile(context.Background(), tmpDir, "http://tracker.example.com/announce")
	if err != nil {
		// We expect this to fail without mktorrent installed
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("mktorrent not installed, skipping torrent creation test")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify torrent file was created
	if _, err := os.Stat(torrentPath); os.IsNotExist(err) {
		t.Error("torrent file was not created")
	}
}

func TestUploadCommand_ValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		meta    *Metadata
		wantErr bool
		errMsg  string
	}{
		{
			name: "all fields present",
			meta: &Metadata{
				Title:       "Album Title",
				Year:        2013,
				Format:      "FLAC",
				Encoding:    "Lossless",
				Media:       "CD",
				Tags:        []string{"classical"},
				Artists:     []ArtistCredit{{Name: "Artist", Role: "artists"}},
				Description: "Description",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			meta: &Metadata{
				Year:     2013,
				Format:   "FLAC",
				Encoding: "Lossless",
				Media:    "CD",
				Tags:     []string{"classical"},
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "missing tags",
			meta: &Metadata{
				Title:    "Album",
				Year:     2013,
				Format:   "FLAC",
				Encoding: "Lossless",
				Media:    "CD",
				Tags:     []string{}, // Empty
			},
			wantErr: true,
			errMsg:  "tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &UploadCommand{}
			err := cmd.validateRequiredFields(tt.meta)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequiredFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error should contain %q, got %v", tt.errMsg, err)
			}
		})
	}
}

// Test for rate limiter integration with existing discogs rate limiter
func TestRateLimiter_Integration(t *testing.T) {
	limiter := ratelimit.NewRateLimiter(2, 2*time.Second) // 2 tokens, refill every 2 seconds

	ctx := context.Background()

	// Should allow first two requests immediately
	for i := 0; i < 2; i++ {
		start := time.Now()
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		elapsed := time.Since(start)
		if elapsed > 100*time.Millisecond {
			t.Errorf("request %d took too long: %v", i, elapsed)
		}
		limiter.OnResponse() // Simulate response received
	}

	// Third request should wait
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("third request failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 1*time.Second {
		t.Errorf("third request didn't wait long enough: %v", elapsed)
	}
}
