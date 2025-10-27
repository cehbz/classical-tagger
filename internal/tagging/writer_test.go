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
		Name     string
		Track    *domain.Track
		Album    *domain.Album
		WantTags map[string]string
	}{
		{
			Name: "single composer, single performer",
			Track: func() *domain.Track {
				composer := domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}
				performer := domain.Artist{Name: "Glenn Gould", Role: domain.RoleSoloist}
				return &domain.Track{
					Disc:    1,
					Track:   1,
					Title:   "Goldberg Variations, BWV 988: Aria",
					Artists: []domain.Artist{composer, performer},
				}
			}(),
			Album: func() *domain.Album {
				return &domain.Album{Title: "Goldberg Variations", OriginalYear: 1981}
			}(),
			WantTags: map[string]string{
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
			Name: "multiple performers with roles",
			Track: func() *domain.Track {
				composer := domain.Artist{Name: "Johannes Brahms", Role: domain.RoleComposer}
				soloist := domain.Artist{Name: "Anne-Sophie Mutter", Role: domain.RoleSoloist}
				ensemble := domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}
				conductor := domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}
				return &domain.Track{
					Disc:    1,
					Track:   1,
					Title:   "Violin Concerto in D major, Op. 77: I. Allegro non troppo",
					Artists: []domain.Artist{composer, soloist, ensemble, conductor},
				}
			}(),
			Album: func() *domain.Album {
				return &domain.Album{Title: "Brahms: Violin Concerto", OriginalYear: 1980}
			}(),
			WantTags: map[string]string{
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
			Name: "with edition info",
			Track: func() *domain.Track {
				composer := domain.Artist{Name: "Felix Mendelssohn Bartholdy", Role: domain.RoleComposer}
				ensemble := domain.Artist{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble}
				return &domain.Track{
					Disc:    1,
					Track:   1,
					Title:   "Frohlocket, Op. 79/1",
					Artists: []domain.Artist{composer, ensemble},
				}
			}(),
			Album: func() *domain.Album {
				return &domain.Album{
					Title:        "Christmas Music",
					OriginalYear: 2013,
					Edition: &domain.Edition{
						Label:         "test label",
						Year:          2013,
						CatalogNumber: "HMC902170",
					},
				}
			}(),
			WantTags: map[string]string{
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
			Name: "original recording remastered - different years",
			Track: func() *domain.Track {
				composer := domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}
				performer := domain.Artist{Name: "Glenn Gould", Role: domain.RoleSoloist}
				return &domain.Track{
					Disc:    1,
					Track:   1,
					Title:   "Goldberg Variations, BWV 988: Aria",
					Artists: []domain.Artist{composer, performer},
				}
			}(),
			Album: func() *domain.Album {
				return &domain.Album{
					Title:        "Goldberg Variations",
					OriginalYear: 1955, // Original recording
					Edition: &domain.Edition{ // Remaster edition
						Label:         "Sony Classical",
						Year:          1992,
						CatalogNumber: "SK 52594",
					},
				}
			}(),
			WantTags: map[string]string{
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
		t.Run(tt.Name, func(t *testing.T) {
			tags := MetadataToVorbisComment(tt.Track, tt.Album)

			for key, want := range tt.WantTags {
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
				if _, expected := tt.WantTags[key]; !expected {
					t.Errorf("MetadataToVorbisComment() unexpected tag %q = %q", key, tags[key])
				}
			}
		})
	}
}
