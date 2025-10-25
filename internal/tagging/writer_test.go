package tagging

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// NOTE: The main integration tests are in writer_integration_test.go
// This file contains the unit tests for MetadataToVorbisComment

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
				edition, _ := domain.NewEdition("test label", 2013)
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
				"LABEL":         "test label",
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

			// Check no extra tags (except ALBUMARTIST which is optional)
			for key := range tags {
				if key == "ALBUMARTIST" {
					continue // Optional tag
				}
				if _, expected := tt.wantTags[key]; !expected {
					t.Errorf("MetadataToVorbisComment() unexpected tag %q = %q", key, tags[key])
				}
			}
		})
	}
}
