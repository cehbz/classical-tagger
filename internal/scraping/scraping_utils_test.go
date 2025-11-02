package scraping

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "standard entities",
			Input: "Bach &amp; Beethoven",
			Want:  "Bach & Beethoven",
		},
		{
			Name:  "quotes",
			Input: "&quot;Goldberg Variations&quot;",
			Want:  `"Goldberg Variations"`,
		},
		{
			Name:  "apostrophe",
			Input: "Bach&#039;s Suite",
			Want:  "Bach's Suite",
		},
		{
			Name:  "malformed UTF-8 - Noël",
			Input: "NoÃ«l",
			Want:  "Noël",
		},
		{
			Name:  "malformed UTF-8 - umlaut",
			Input: "MÃ¼ller",
			Want:  "Müller",
		},
		{
			Name:  "no entities",
			Input: "Plain Text",
			Want:  "Plain Text",
		},
		{
			Name:  "mixed",
			Input: "NoÃ«l &amp; MÃ¼ller",
			Want:  "Noël & Müller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := decodeHTMLEntities(tt.Input)
			if got != tt.Want {
				t.Errorf("decodeHTMLEntities(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "simple tags",
			Input: "<b>Bold</b> and <i>italic</i>",
			Want:  "Bold and italic",
		},
		{
			Name:  "nested tags",
			Input: "<div><span>Text</span></div>",
			Want:  "Text",
		},
		{
			Name:  "tags with attributes",
			Input: `<a href="link">Link</a>`,
			Want:  "Link",
		},
		{
			Name:  "no tags",
			Input: "Plain text",
			Want:  "Plain text",
		},
		{
			Name:  "empty tags",
			Input: "Text <br/> more text",
			Want:  "Text  more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := stripHTMLTags(tt.Input)
			if got != tt.Want {
				t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "multiple spaces",
			Input: "Too    many   spaces",
			Want:  "Too many spaces",
		},
		{
			Name:  "leading and trailing",
			Input: "  trim me  ",
			Want:  "trim me",
		},
		{
			Name:  "tabs and newlines",
			Input: "text\t\twith\ntabs\nand\nnewlines",
			Want:  "text with tabs and newlines",
		},
		{
			Name:  "already clean",
			Input: "clean text",
			Want:  "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := cleanWhitespace(tt.Input)
			if got != tt.Want {
				t.Errorf("cleanWhitespace(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "all caps",
			Input: "GOLDBERG VARIATIONS",
			Want:  "Goldberg Variations",
		},
		{
			Name:  "with articles",
			Input: "THE ART OF FUGUE",
			Want:  "The Art of Fugue",
		},
		{
			Name:  "with prepositions",
			Input: "CONCERTO IN D MAJOR",
			Want:  "Concerto in D Major",
		},
		{
			Name:  "already title case",
			Input: "Symphony No. 5",
			Want:  "Symphony No. 5",
		},
		{
			Name:  "with de/la/von",
			Input: "MUSIC OF LA RUE",
			Want:  "Music of la Rue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := toTitleCase(tt.Input)
			if got != tt.Want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "tabs to spaces",
			Input: "text\twith\ttabs",
			Want:  "text with tabs",
		},
		{
			Name:  "newlines to spaces",
			Input: "line1\nline2\nline3",
			Want:  "line1 line2 line3",
		},
		{
			Name:  "non-breaking spaces",
			Input: "text\u00A0with\u00A0nbsp",
			Want:  "text with nbsp",
		},
		{
			Name:  "mixed whitespace",
			Input: "text\t\n\r  with   mixed",
			Want:  "text with mixed",
		},
		{
			Name:  "multiple spaces",
			Input: "Too    many   spaces",
			Want:  "Too many spaces",
		},
		{
			Name:  "leading and trailing",
			Input: "  trim me  ",
			Want:  "trim me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := normalizeWhitespace(tt.Input)
			if got != tt.Want {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestSanitizeText(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "complete sanitization",
			Input: "<b>NoÃ«l</b> &amp; <i>MÃ¼ller</i>",
			Want:  "Noël & Müller",
		},
		{
			Name:  "html with entities and whitespace",
			Input: "  <div>Text   with   &nbsp;  spaces</div>  ",
			Want:  "Text with spaces",
		},
		{
			Name:  "complex example",
			Input: "<h1>Bach&#039;s   Goldberg\n\nVariations</h1>",
			Want:  "Bach's Goldberg Variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := sanitizeText(tt.Input)
			if got != tt.Want {
				t.Errorf("sanitizeText(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestCleanHTMLEntities_LegacyAlias(t *testing.T) {
	// Test that legacy function still works
	input := "NoÃ«l &amp; Christmas"
	want := "Noël & Christmas"
	got := cleanHTMLEntities(input)

	if got != want {
		t.Errorf("cleanHTMLEntities(%q) = %q, want %q", input, got, want)
	}
}

func TestRemoveArtistsFromTracks(t *testing.T) {
	tests := []struct {
		Name            string
		Tracks          []*domain.Track
		ArtistsToRemove []domain.Artist
		Want            []*domain.Track
	}{
		{
			Name: "remove ensemble from all tracks",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
				{
					Disc:  1,
					Track: 2,
					Title: "Track 2",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
			},
			ArtistsToRemove: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Want: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
				{
					Disc:  1,
					Track: 2,
					Title: "Track 2",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
			},
		},
		{
			Name: "remove multiple artists",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
			},
			ArtistsToRemove: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
			Want: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
					},
				},
			},
		},
		{
			Name: "match by name AND role - same name different role not removed",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
						{Name: "Bach", Role: domain.RoleConductor}, // Same name, different role
					},
				},
			},
			ArtistsToRemove: []domain.Artist{
				{Name: "Bach", Role: domain.RoleComposer},
			},
			Want: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleConductor}, // Should remain
					},
				},
			},
		},
		{
			Name: "no artists to remove",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
					},
				},
			},
			ArtistsToRemove: []domain.Artist{},
			Want: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
					},
				},
			},
		},
		{
			Name: "artist not in tracks",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
					},
				},
			},
			ArtistsToRemove: []domain.Artist{
				{Name: "Mozart", Role: domain.RoleComposer}, // Not in tracks
			},
			Want: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Bach", Role: domain.RoleComposer},
					},
				},
			},
		},
		{
			Name: "remove all artists from track",
			Tracks: []*domain.Track{
				{
					Disc:  1,
					Track: 1,
					Title: "Track 1",
					Artists: []domain.Artist{
						{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						{Name: "Herbert von Karajan", Role: domain.RoleConductor},
					},
				},
			},
			ArtistsToRemove: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
			Want: []*domain.Track{
				{
					Disc:    1,
					Track:   1,
					Title:   "Track 1",
					Artists: []domain.Artist{}, // All removed
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Make a copy of tracks to avoid modifying the original test data
			tracksCopy := make([]*domain.Track, len(tt.Tracks))
			for i, track := range tt.Tracks {
				artistsCopy := make([]domain.Artist, len(track.Artists))
				copy(artistsCopy, track.Artists)
				tracksCopy[i] = &domain.Track{
					Disc:    track.Disc,
					Track:   track.Track,
					Title:   track.Title,
					Artists: artistsCopy,
				}
			}

			removeArtistsFromTracks(tracksCopy, tt.ArtistsToRemove)

			// Verify results
			if len(tracksCopy) != len(tt.Want) {
				t.Fatalf("removeArtistsFromTracks() returned %d tracks, want %d", len(tracksCopy), len(tt.Want))
			}

			for i, gotTrack := range tracksCopy {
				wantTrack := tt.Want[i]
				if len(gotTrack.Artists) != len(wantTrack.Artists) {
					t.Errorf("Track %d: got %d artists, want %d", i+1, len(gotTrack.Artists), len(wantTrack.Artists))
					continue
				}

				for j, gotArtist := range gotTrack.Artists {
					wantArtist := wantTrack.Artists[j]
					if gotArtist.Name != wantArtist.Name || gotArtist.Role != wantArtist.Role {
						t.Errorf("Track %d, Artist %d: got %+v, want %+v", i+1, j+1, gotArtist, wantArtist)
					}
				}
			}
		})
	}
}

func TestMergePerformers(t *testing.T) {
	tests := []struct {
		Name       string
		Existing   []domain.Artist
		Additional []domain.Artist
		Want       []domain.Artist
	}{
		{
			Name: "no duplicates - simple merge",
			Existing: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Additional: []domain.Artist{
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
			Want: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
		},
		{
			Name: "duplicate by name and role - should be removed",
			Existing: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			},
			Additional: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble}, // Duplicate
				{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
			},
			Want: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
				{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
			},
		},
		{
			Name: "same name different role - should both be included",
			Existing: []domain.Artist{
				{Name: "Bach", Role: domain.RoleComposer},
			},
			Additional: []domain.Artist{
				{Name: "Bach", Role: domain.RoleConductor}, // Same name, different role
			},
			Want: []domain.Artist{
				{Name: "Bach", Role: domain.RoleComposer},
				{Name: "Bach", Role: domain.RoleConductor},
			},
		},
		{
			Name:     "empty existing",
			Existing: []domain.Artist{},
			Additional: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Want: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
		},
		{
			Name: "empty additional",
			Existing: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
			Additional: []domain.Artist{},
			Want: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
			},
		},
		{
			Name:       "both empty",
			Existing:   []domain.Artist{},
			Additional: []domain.Artist{},
			Want:       []domain.Artist{},
		},
		{
			Name: "complex merge with multiple duplicates",
			Existing: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
				{Name: "Glenn Gould", Role: domain.RoleSoloist},
			},
			Additional: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},  // Duplicate
				{Name: "Herbert von Karajan", Role: domain.RoleConductor}, // Duplicate
				{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},  // New
			},
			Want: []domain.Artist{
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
				{Name: "Glenn Gould", Role: domain.RoleSoloist},
				{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := mergePerformers(tt.Existing, tt.Additional)

			if len(got) != len(tt.Want) {
				t.Fatalf("mergePerformers() returned %d artists, want %d", len(got), len(tt.Want))
			}

			// Create maps for comparison (order doesn't matter for correctness, but we'll check both)
			gotMap := make(map[string]map[domain.Role]bool)
			for _, artist := range got {
				if gotMap[artist.Name] == nil {
					gotMap[artist.Name] = make(map[domain.Role]bool)
				}
				gotMap[artist.Name][artist.Role] = true
			}

			wantMap := make(map[string]map[domain.Role]bool)
			for _, artist := range tt.Want {
				if wantMap[artist.Name] == nil {
					wantMap[artist.Name] = make(map[domain.Role]bool)
				}
				wantMap[artist.Name][artist.Role] = true
			}

			// Verify all wanted artists are present
			for _, wantArtist := range tt.Want {
				if gotMap[wantArtist.Name] == nil || !gotMap[wantArtist.Name][wantArtist.Role] {
					t.Errorf("mergePerformers() missing artist: %+v", wantArtist)
				}
			}

			// Verify no extra artists
			for _, gotArtist := range got {
				if wantMap[gotArtist.Name] == nil || !wantMap[gotArtist.Name][gotArtist.Role] {
					t.Errorf("mergePerformers() has extra artist: %+v", gotArtist)
				}
			}
		})
	}
}
