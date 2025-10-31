package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerTagRequired(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name         string
		Actual       *domain.Album
		WantPass     bool
		WantErrors   int
		WantWarnings int
		Expect       CaseExpectation
	}{
		{
			Name:     "valid - full composer name",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:     "valid - composer with initials",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "J.S. Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:     "valid - composer with spaced initials",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "J. S. Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:     "valid - two-word name",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Johann Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:     "valid - composer with surname prefix",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Wolfgang Amadeus Mozart", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:       "invalid - last name only (ambiguous)",
			Actual:     NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Bach", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass:   false,
			WantErrors: 1,
			Expect:     CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name: "invalid - missing composer",
			Actual: func() *domain.Album {
				ensemble := domain.Artist{Name: "Vienna Phil", Role: domain.RoleEnsemble}
				track := domain.Track{Disc: 1, Track: 1, Title: "Symphony", Artists: []domain.Artist{ensemble}}
				return &domain.Album{Title: "Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track}}
			}(),
			WantPass:   false,
			WantErrors: 1,
			Expect:     CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name: "multiple tracks, one missing composer",
			Actual: func() *domain.Album {
				composer := domain.Artist{Name: "Ludwig van Beethoven", Role: domain.RoleComposer}
				ensemble := domain.Artist{Name: "Vienna Phil", Role: domain.RoleEnsemble}

				track1 := domain.Track{Disc: 1, Track: 1, Title: "Symphony No. 1", Artists: []domain.Artist{composer, ensemble}}
				track2 := domain.Track{Disc: 1, Track: 2, Title: "Symphony No. 2", Artists: []domain.Artist{ensemble}}
				track3 := domain.Track{Disc: 1, Track: 3, Title: "Symphony No. 3", Artists: []domain.Artist{composer, ensemble}}

				return &domain.Album{Title: "Beethoven Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track1, &track2, &track3}}
			}(),
			WantPass:   false,
			WantErrors: 1,
			Expect:     CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}, {Errors: 1, Warnings: 0, Info: 0}, {Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name: "multiple tracks, some ambiguous names",
			Actual: func() *domain.Album {
				composer1 := domain.Artist{Name: "Johann Sebastian Bach", Role: domain.RoleComposer}
				composer2 := domain.Artist{Name: "Bach", Role: domain.RoleComposer} // Ambiguous
				ensemble := domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble}

				track1 := domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{composer1, ensemble}}
				track2 := domain.Track{Disc: 1, Track: 2, Title: "Work 2", Artists: []domain.Artist{composer2, ensemble}}

				return &domain.Album{Title: "Bach Works", OriginalYear: 1963, Tracks: []*domain.Track{&track1, &track2}}
			}(),
			WantPass:   false,
			WantErrors: 1,
			Expect:     CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}, {Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:     "edge case - Beethoven, Ludwig van (reversed format)",
			Actual:   NewAlbum().WithTitle("Beethoven Symphonies").ClearTracks().AddTrack().WithTitle("Symphony No. 5").ClearArtists().WithArtists(domain.Artist{Name: "Beethoven, Ludwig van", Role: domain.RoleComposer}, domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}, domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}).Build().Build(),
			WantPass: true,
			Expect:   CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
	}

	for _, tt := range tests {
		if tt.Expect != nil {
			if len(tt.Expect) != len(tt.Actual.Tracks) {
				t.Fatalf("Expect length %d does not match tracks %d for case %s", len(tt.Expect), len(tt.Actual.Tracks), tt.Name)
			}
		}
		for i, track := range tt.Actual.Tracks {
			name := tt.Name
			if len(tt.Actual.Tracks) > 1 {
				name = name + "/track#" + string(rune('1'+i))
			}
			t.Run(name, func(t *testing.T) {
				result := rules.ComposerTagRequired(track, nil, nil, nil)

				errors, warnings, info := 0, 0, 0
				for _, issue := range result.Issues {
					switch issue.Level {
					case domain.LevelError:
						errors++
					case domain.LevelWarning:
						warnings++
					case domain.LevelInfo:
						info++
					}
				}

				if tt.Expect != nil {
					exp := tt.Expect[i]
					if errors != exp.Errors {
						t.Errorf("Errors = %d, want %d", errors, exp.Errors)
					}
					if warnings != exp.Warnings {
						t.Errorf("Warnings = %d, want %d", warnings, exp.Warnings)
					}
					if info != exp.Info {
						t.Errorf("Info = %d, want %d", info, exp.Info)
					}
				}
			})
		}
	}
}
