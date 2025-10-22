package tagging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestFLACWriter_WriteTrack tests writing a single track's metadata to a new file.
func TestFLACWriter_WriteTrack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup: Create a test FLAC file (we'll need a minimal valid FLAC)
	// For now, we'll test the interface
	t.Skip("requires test FLAC file fixture")

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.flac")
	destPath := filepath.Join(tmpDir, "dest.flac")

	// Create test track
	composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)
	track, _ := domain.NewTrack(1, 1, "Goldberg Variations, BWV 988: Aria", []domain.Artist{composer, ensemble})

	writer := NewFLACWriter()
	album, _ := domain.NewAlbum("Goldberg Variations", 1981)
	edition, _ := domain.NewEdition("Sony Classical", 1992)
	edition = edition.WithCatalogNumber("SK 52594")
	album = album.WithEdition(edition)
	err := writer.WriteTrack(sourcePath, destPath, track, album)
	if err != nil {
		t.Fatalf("WriteTrack() error = %v", err)
	}

	// Verify dest file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}
}

// TestFLACWriter_PreservesAudio tests that audio data is copied bit-perfect.
func TestFLACWriter_PreservesAudio(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Skip("requires test FLAC file fixture and audio comparison")
}

// TestMetadataToVorbisComment tests conversion from domain Track to Vorbis comment tags.
func TestMetadataToVorbisComment(t *testing.T) {
	tests := []struct {
		name     string
		track    *domain.Track
		album    *domain.Album
		wantTags map[string]string
	}{
		{
			name: "single composer, single performer",
			track: func() *domain.Track {
				composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
				performer, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)
				track, _ := domain.NewTrack(1, 1, "Goldberg Variations, BWV 988: Aria",
					[]domain.Artist{composer, performer})
				return track
			}(),
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Goldberg Variations", 1981)
				return album
			}(),
			wantTags: map[string]string{
				"COMPOSER":     "Johann Sebastian Bach",
				"ARTIST":       "Glenn Gould",
				"PERFORMER":    "Glenn Gould",
				"TITLE":        "Goldberg Variations, BWV 988: Aria",
				"ALBUM":        "Goldberg Variations",
				"TRACKNUMBER":  "1",
				"DISCNUMBER":   "1",
				"ORIGINALDATE": "1981",
			},
		},
		{
			name: "multiple performers with roles",
			track: func() *domain.Track {
				composer, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
				soloist, _ := domain.NewArtist("Anne-Sophie Mutter", domain.RoleSoloist)
				ensemble, _ := domain.NewArtist("Berlin Philharmonic", domain.RoleEnsemble)
				conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)
				track, _ := domain.NewTrack(1, 1, "Violin Concerto in D major, Op. 77: I. Allegro non troppo",
					[]domain.Artist{composer, soloist, ensemble, conductor})
				return track
			}(),
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Brahms: Violin Concerto", 1980)
				return album
			}(),
			wantTags: map[string]string{
				"COMPOSER":     "Johannes Brahms",
				"ARTIST":       "Anne-Sophie Mutter, Berlin Philharmonic, Herbert von Karajan",
				"PERFORMER":    "Anne-Sophie Mutter",
				"ENSEMBLE":     "Berlin Philharmonic",
				"CONDUCTOR":    "Herbert von Karajan",
				"TITLE":        "Violin Concerto in D major, Op. 77: I. Allegro non troppo",
				"ALBUM":        "Brahms: Violin Concerto",
				"TRACKNUMBER":  "1",
				"DISCNUMBER":   "1",
				"ORIGINALDATE": "1980",
			},
		},
		{
			name: "with edition info",
			track: func() *domain.Track {
				composer, _ := domain.NewArtist("Felix Mendelssohn Bartholdy", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("RIAS Kammerchor", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Frohlocket, Op. 79/1",
					[]domain.Artist{composer, ensemble})
				return track
			}(),
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Christmas Music", 2013)
				edition, _ := domain.NewEdition("harmonia mundi", 2013)
				edition = edition.WithCatalogNumber("HMC902170")
				album = album.WithEdition(edition)
				return album
			}(),
			wantTags: map[string]string{
				"COMPOSER":      "Felix Mendelssohn Bartholdy",
				"ARTIST":        "RIAS Kammerchor",
				"ENSEMBLE":      "RIAS Kammerchor",
				"TITLE":         "Frohlocket, Op. 79/1",
				"ALBUM":         "Christmas Music",
				"TRACKNUMBER":   "1",
				"DISCNUMBER":    "1",
				"ORIGINALDATE":  "2013",
				"DATE":          "2013", // Edition year
				"LABEL":         "harmonia mundi",
				"CATALOGNUMBER": "HMC902170",
			},
		},
		{
			name: "original recording remastered - different years",
			track: func() *domain.Track {
				composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
				performer, _ := domain.NewArtist("Glenn Gould", domain.RoleSoloist)
				track, _ := domain.NewTrack(1, 1, "Goldberg Variations, BWV 988: Aria",
					[]domain.Artist{composer, performer})
				return track
			}(),
			album: func() *domain.Album {
				album, _ := domain.NewAlbum("Goldberg Variations", 1955) // Original recording
				edition, _ := domain.NewEdition("Sony Classical", 1992)  // Remaster edition
				edition = edition.WithCatalogNumber("SK 52594")
				album = album.WithEdition(edition)
				return album
			}(),
			wantTags: map[string]string{
				"COMPOSER":      "Johann Sebastian Bach",
				"ARTIST":        "Glenn Gould",
				"PERFORMER":     "Glenn Gould",
				"TITLE":         "Goldberg Variations, BWV 988: Aria",
				"ALBUM":         "Goldberg Variations",
				"TRACKNUMBER":   "1",
				"DISCNUMBER":    "1",
				"ORIGINALDATE":  "1955", // Original recording year
				"DATE":          "1992", // Remaster/edition year
				"LABEL":         "Sony Classical",
				"CATALOGNUMBER": "SK 52594",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := MetadataToVorbisComment(tt.track, tt.album)

			for key, want := range tt.wantTags {
				got, exists := tags[key]
				if !exists {
					t.Errorf("MetadataToVorbisComment() missing tag %q", key)
					continue
				}
				if got != want {
					t.Errorf("MetadataToVorbisComment() tag %q = %q, want %q", key, got, want)
				}
			}

			// Check no extra tags
			for key := range tags {
				if _, expected := tt.wantTags[key]; !expected {
					t.Errorf("MetadataToVorbisComment() unexpected tag %q = %q", key, tags[key])
				}
			}
		})
	}
}

// TestFormatArtists tests formatting multiple artists according to classical music rules.
func TestFormatArtists(t *testing.T) {
	tests := []struct {
		name    string
		artists []domain.Artist
		want    string
	}{
		{
			name: "single soloist",
			artists: []domain.Artist{
				func() domain.Artist { a, _ := domain.NewArtist("Martha Argerich", domain.RoleSoloist); return a }(),
			},
			want: "Martha Argerich",
		},
		{
			name: "soloist, ensemble, conductor",
			artists: []domain.Artist{
				func() domain.Artist { a, _ := domain.NewArtist("Martha Argerich", domain.RoleSoloist); return a }(),
				func() domain.Artist { a, _ := domain.NewArtist("Berlin Philharmonic", domain.RoleEnsemble); return a }(),
				func() domain.Artist { a, _ := domain.NewArtist("Claudio Abbado", domain.RoleConductor); return a }(),
			},
			want: "Martha Argerich, Berlin Philharmonic, Claudio Abbado",
		},
		{
			name: "multiple soloists, ensemble, conductor",
			artists: []domain.Artist{
				func() domain.Artist { a, _ := domain.NewArtist("Anne-Sophie Mutter", domain.RoleSoloist); return a }(),
				func() domain.Artist { a, _ := domain.NewArtist("Yo-Yo Ma", domain.RoleSoloist); return a }(),
				func() domain.Artist {
					a, _ := domain.NewArtist("Boston Symphony Orchestra", domain.RoleEnsemble)
					return a
				}(),
				func() domain.Artist { a, _ := domain.NewArtist("Seiji Ozawa", domain.RoleConductor); return a }(),
			},
			want: "Anne-Sophie Mutter, Yo-Yo Ma, Boston Symphony Orchestra, Seiji Ozawa",
		},
		{
			name: "ensemble only",
			artists: []domain.Artist{
				func() domain.Artist { a, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble); return a }(),
			},
			want: "Vienna Philharmonic",
		},
		{
			name: "composer filtered out",
			artists: []domain.Artist{
				func() domain.Artist { a, _ := domain.NewArtist("Mozart", domain.RoleComposer); return a }(),
				func() domain.Artist { a, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble); return a }(),
			},
			want: "Vienna Philharmonic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatArtists(tt.artists)
			if got != tt.want {
				t.Errorf("FormatArtists() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDetermineAlbumArtist tests finding universal performers across tracks.
func TestDetermineAlbumArtist(t *testing.T) {
	tests := []struct {
		name             string
		setupAlbum       func() *domain.Album
		wantAlbumArtist  string
		wantUniversalLen int
	}{
		{
			name: "same ensemble and conductor throughout",
			setupAlbum: func() *domain.Album {
				album, _ := domain.NewAlbum("Test Album", 2020)

				composer1, _ := domain.NewArtist("Composer 1", domain.RoleComposer)
				composer2, _ := domain.NewArtist("Composer 2", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Test Ensemble", domain.RoleEnsemble)
				conductor, _ := domain.NewArtist("Test Conductor", domain.RoleConductor)

				// Track 1
				track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{composer1, ensemble, conductor})
				album.AddTrack(track1)

				// Track 2 - same performers, different composer
				track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{composer2, ensemble, conductor})
				album.AddTrack(track2)

				return album
			},
			wantAlbumArtist:  "Test Ensemble, Test Conductor",
			wantUniversalLen: 2,
		},
		{
			name: "no universal performers",
			setupAlbum: func() *domain.Album {
				album, _ := domain.NewAlbum("Test Album", 2020)

				composer, _ := domain.NewArtist("Composer", domain.RoleComposer)
				soloist1, _ := domain.NewArtist("Soloist 1", domain.RoleSoloist)
				soloist2, _ := domain.NewArtist("Soloist 2", domain.RoleSoloist)

				// Track 1 - Soloist 1
				track1, _ := domain.NewTrack(1, 1, "Track 1", []domain.Artist{composer, soloist1})
				album.AddTrack(track1)

				// Track 2 - Soloist 2 (different)
				track2, _ := domain.NewTrack(1, 2, "Track 2", []domain.Artist{composer, soloist2})
				album.AddTrack(track2)

				return album
			},
			wantAlbumArtist:  "",
			wantUniversalLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			album := tt.setupAlbum()
			albumArtist, universal := DetermineAlbumArtist(album)

			if albumArtist != tt.wantAlbumArtist {
				t.Errorf("DetermineAlbumArtist() albumArtist = %q, want %q", albumArtist, tt.wantAlbumArtist)
			}

			if len(universal) != tt.wantUniversalLen {
				t.Errorf("DetermineAlbumArtist() universal count = %d, want %d", len(universal), tt.wantUniversalLen)
			}
		})
	}
}
