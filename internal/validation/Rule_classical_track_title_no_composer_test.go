package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerNotInTitle(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name   string
		Actual *domain.Album
		Expect CaseExpectation
	}{
		{
			Name:   "valid - no composer in title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 5 in C Minor, Op. 67").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer last name in title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Beethoven: Symphony No. 5").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer appended to title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Brandenburg Concerto No. 1 - Bach").ClearArtists().WithArtist("Johann Sebastian Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - composer with initials",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Brandenburg Concerto No. 1").ClearArtists().WithArtist("J.S. Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer surname with initials in title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("J.S. Bach: Brandenburg Concerto No. 1").ClearArtists().WithArtist("J.S. Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - exception for work title containing composer",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Variations on a Theme by Haydn").ClearArtists().WithArtist("Johannes Brahms", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - 'after composer' is part of work title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Concerto after Vivaldi").ClearArtists().WithArtist("Igor Stravinsky", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer in parentheses",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 40 (Mozart)").ClearArtists().WithArtist("Wolfgang Amadeus Mozart", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - different word containing composer name",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Bacharach Suite").ClearArtists().WithArtist("Johann Sebastian Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}}, // "Bacharach" is a different word, not "Bach"
		},
		{
			Name:   "invalid - composer with compound last name",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Beethoven: Piano Sonata").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - composer with compound last name, not in title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Piano Sonata No. 14").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - reversed name format",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Symphony No. 9").ClearArtists().WithArtist("Beethoven, Ludwig van", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - reversed name format, composer in title",
			Actual: NewAlbum().ClearTracks().AddTrack().WithTitle("Beethoven: Symphony No. 9").ClearArtists().WithArtist("Beethoven, Ludwig van", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name: "multiple tracks, some with composer in title",
			Actual: func() *domain.Album {
				composer := domain.Artist{Name: "Johannes Brahms", Role: domain.RoleComposer}
				ensemble := domain.Artist{Name: "Vienna Phil", Role: domain.RoleEnsemble}
				artists := []domain.Artist{composer, ensemble}

				track1 := domain.Track{Disc: 1, Track: 1, Title: "Symphony No. 1", Artists: artists}
				track2 := domain.Track{Disc: 1, Track: 2, Title: "Brahms: Symphony No. 2", Artists: artists}
				track3 := domain.Track{Disc: 1, Track: 3, Title: "Symphony No. 3", Artists: artists}

				return &domain.Album{Title: "Brahms Symphonies", OriginalYear: 1963, Tracks: []*domain.Track{&track1, &track2, &track3}}
			}(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}, {Errors: 1, Warnings: 0, Info: 0}, {Errors: 0, Warnings: 0, Info: 0}},
		},
	}

	for _, tt := range tests {
		if tt.Expect != nil && len(tt.Expect) != len(tt.Actual.Tracks) {
			t.Fatalf("Expect length %d does not match tracks %d for case %s", len(tt.Expect), len(tt.Actual.Tracks), tt.Name)
		}
		for i, tr := range tt.Actual.Tracks {
			track := tr
			name := tt.Name
			if len(tt.Actual.Tracks) > 1 {
				name = name + "/track#" + string(rune('1'+i))
			}
			t.Run(name, func(t *testing.T) {
				result := rules.ComposerNotInTitle(track, nil, nil, nil)

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
