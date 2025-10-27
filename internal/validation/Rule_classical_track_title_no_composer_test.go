package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerNotInTitle(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name:     "valid - no composer in title",
			Actual:   buildAlbumWithTitle("Ludwig van Beethoven", "Symphony No. 5 in C Minor, Op. 67"),
			WantPass: true,
		},
		{
			Name:       "invalid - composer last name in title",
			Actual:     buildAlbumWithTitle("Ludwig van Beethoven", "Beethoven: Symphony No. 5"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - composer appended to title",
			Actual:     buildAlbumWithTitle("Johann Sebastian Bach", "Brandenburg Concerto No. 1 - Bach"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:     "valid - composer with initials",
			Actual:   buildAlbumWithTitle("J.S. Bach", "Brandenburg Concerto No. 1"),
			WantPass: true,
		},
		{
			Name:       "invalid - composer surname with initials in title",
			Actual:     buildAlbumWithTitle("J.S. Bach", "J.S. Bach: Brandenburg Concerto No. 1"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:     "valid - exception for work title containing composer",
			Actual:   buildAlbumWithTitle("Johannes Brahms", "Variations on a Theme by Haydn"),
			WantPass: true,
		},
		{
			Name:     "valid - 'after composer' is part of work title",
			Actual:   buildAlbumWithTitle("Igor Stravinsky", "Concerto after Vivaldi"),
			WantPass: true,
		},
		{
			Name:       "invalid - composer in parentheses",
			Actual:     buildAlbumWithTitle("Wolfgang Amadeus Mozart", "Symphony No. 40 (Mozart)"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:     "valid - different word containing composer name",
			Actual:   buildAlbumWithTitle("Johann Sebastian Bach", "Bacharach Suite"),
			WantPass: true, // "Bacharach" is a different word, not "Bach"
		},
		{
			Name:       "invalid - composer with compound last name",
			Actual:     buildAlbumWithTitle("Ludwig van Beethoven", "Beethoven: Piano Sonata"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:     "valid - composer with compound last name, not in title",
			Actual:   buildAlbumWithTitle("Ludwig van Beethoven", "Piano Sonata No. 14"),
			WantPass: true,
		},
		{
			Name:     "valid - reversed name format",
			Actual:   buildAlbumWithTitle("Beethoven, Ludwig van", "Symphony No. 9"),
			WantPass: true,
		},
		{
			Name:       "invalid - reversed name format, composer in title",
			Actual:     buildAlbumWithTitle("Beethoven, Ludwig van", "Beethoven: Symphony No. 9"),
			WantPass:   false,
			WantIssues: 1,
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
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.ComposerNotInTitle(tt.Actual, tt.Actual)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass && len(result.Issues) != tt.WantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues), tt.WantIssues)
				for _, issue := range result.Issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

// Helper function to build an album with specific composer and title
func buildAlbumWithTitle(composerName, trackTitle string) *domain.Album {
	composer := domain.Artist{Name: composerName, Role: domain.RoleComposer}
	ensemble := domain.Artist{Name: "Vienna Philharmonic", Role: domain.RoleEnsemble}
	conductor := domain.Artist{Name: "Herbert von Karajan", Role: domain.RoleConductor}

	artists := []domain.Artist{composer, ensemble, conductor}
	track := &domain.Track{Disc: 1, Track: 1, Title: trackTitle, Artists: artists}
	return &domain.Album{Title: "Classical Album", OriginalYear: 1963, Tracks: []*domain.Track{track}}
}

func TestExtractLastNames(t *testing.T) {
	tests := []struct {
		Name         string
		ComposerName string
		Want         []string
	}{
		{"simple name", "Johann Bach", []string{"Bach"}},
		{"full name", "Johann Sebastian Bach", []string{"Bach"}},
		{"with particle", "Ludwig van Beethoven", []string{"van Beethoven"}},
		{"with initials", "J.S. Bach", []string{"Bach"}},
		{"reversed format", "Beethoven, Ludwig van", []string{"Beethoven"}},
		{"compound particle", "Felix Mendelssohn Bartholdy", []string{"Bartholdy"}},
		{"single word", "Bach", []string{"Bach"}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractLastNames(tt.ComposerName)
			if len(got) != len(tt.Want) {
				t.Errorf("extractLastNames() = %v, want %v", got, tt.Want)
				return
			}
			for i := range got {
				if got[i] != tt.Want[i] {
					t.Errorf("extractLastNames()[%d] = %v, want %v", i, got[i], tt.Want[i])
				}
			}
		})
	}
}
