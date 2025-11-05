package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerNotInTitle(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name   string
		Actual *domain.Torrent
		Expect CaseExpectation
	}{
		{
			Name:   "valid - no composer in title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 5 in C Minor, Op. 67").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer last name in title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Beethoven: Symphony No. 5").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer appended to title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Brandenburg Concerto No. 1 - Bach").ClearArtists().WithArtist("Johann Sebastian Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - composer with initials",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Brandenburg Concerto No. 1").ClearArtists().WithArtist("J.S. Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer surname with initials in title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("J.S. Bach: Brandenburg Concerto No. 1").ClearArtists().WithArtist("J.S. Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - exception for work title containing composer",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Variations on a Theme by Haydn").ClearArtists().WithArtist("Johannes Brahms", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - 'after composer' is part of work title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Concerto after Vivaldi").ClearArtists().WithArtist("Igor Stravinsky", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - composer in parentheses",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 40 (Mozart)").ClearArtists().WithArtist("Wolfgang Amadeus Mozart", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - different word containing composer name",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Bacharach Suite").ClearArtists().WithArtist("Johann Sebastian Bach", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}}, // "Bacharach" is a different word, not "Bach"
		},
		{
			Name:   "invalid - composer with compound last name",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Beethoven: Piano Sonata").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - composer with compound last name, not in title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Piano Sonata No. 14").ClearArtists().WithArtist("Ludwig van Beethoven", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "valid - reversed name format",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Symphony No. 9").ClearArtists().WithArtist("Beethoven, Ludwig van", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}},
		},
		{
			Name:   "invalid - reversed name format, composer in title",
			Actual: NewTorrent().ClearTracks().AddTrack().WithTitle("Beethoven: Symphony No. 9").ClearArtists().WithArtist("Beethoven, Ludwig van", domain.RoleComposer).Build().Build(),
			Expect: CaseExpectation{{Errors: 1, Warnings: 0, Info: 0}},
		},
		{
			Name: "multiple tracks, some with composer in title",
			Actual: &domain.Torrent{
				Title:        "Brahms Symphonies",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Symphony No. 1",
						Artists: []domain.Artist{
							{Name: "Johannes Brahms", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 2,
						Title: "Brahms: Symphony No. 2",
						Artists: []domain.Artist{
							{Name: "Johannes Brahms", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 3,
						Title: "Symphony No. 3",
						Artists: []domain.Artist{
							{Name: "Johannes Brahms", Role: domain.RoleComposer},
							{Name: "Vienna Phil", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			Expect: CaseExpectation{{Errors: 0, Warnings: 0, Info: 0}, {Errors: 1, Warnings: 0, Info: 0}, {Errors: 0, Warnings: 0, Info: 0}},
		},
	}

	for _, tt := range tests {
		if tt.Actual == nil {
			t.Fatalf("Actual is nil for case %s", tt.Name)
		}
		tracks := tt.Actual.Tracks()
		if tt.Expect != nil && len(tt.Expect) != len(tracks) {
			t.Fatalf("Expect length %d does not match tracks %d for case %s", len(tt.Expect), len(tracks), tt.Name)
		}
		for i, track := range tracks {
			name := tt.Name
			if len(tracks) > 1 {
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
