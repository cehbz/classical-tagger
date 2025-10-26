package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ComposerNotInTitle(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - no composer in title",
			actual: buildAlbumWithTitle("Ludwig van Beethoven", "Symphony No. 5 in C Minor, Op. 67"),
			wantPass: true,
		},
		{
			name: "invalid - composer last name in title",
			actual: buildAlbumWithTitle("Ludwig van Beethoven", "Beethoven: Symphony No. 5"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "invalid - composer appended to title",
			actual: buildAlbumWithTitle("Johann Sebastian Bach", "Brandenburg Concerto No. 1 - Bach"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "valid - composer with initials",
			actual: buildAlbumWithTitle("J.S. Bach", "Brandenburg Concerto No. 1"),
			wantPass: true,
		},
		{
			name: "invalid - composer surname with initials in title",
			actual: buildAlbumWithTitle("J.S. Bach", "J.S. Bach: Brandenburg Concerto No. 1"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "valid - exception for work title containing composer",
			actual: buildAlbumWithTitle("Johannes Brahms", "Variations on a Theme by Haydn"),
			wantPass: true,
		},
		{
			name: "valid - 'after composer' is part of work title",
			actual: buildAlbumWithTitle("Igor Stravinsky", "Concerto after Vivaldi"),
			wantPass: true,
		},
		{
			name: "invalid - composer in parentheses",
			actual: buildAlbumWithTitle("Wolfgang Amadeus Mozart", "Symphony No. 40 (Mozart)"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "valid - different word containing composer name",
			actual: buildAlbumWithTitle("Johann Sebastian Bach", "Bacharach Suite"),
			wantPass: true, // "Bacharach" is a different word, not "Bach"
		},
		{
			name: "invalid - composer with compound last name",
			actual: buildAlbumWithTitle("Ludwig van Beethoven", "Beethoven: Piano Sonata"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "valid - composer with compound last name, not in title",
			actual: buildAlbumWithTitle("Ludwig van Beethoven", "Piano Sonata No. 14"),
			wantPass: true,
		},
		{
			name: "valid - reversed name format",
			actual: buildAlbumWithTitle("Beethoven, Ludwig van", "Symphony No. 9"),
			wantPass: true,
		},
		{
			name: "invalid - reversed name format, composer in title",
			actual: buildAlbumWithTitle("Beethoven, Ludwig van", "Beethoven: Symphony No. 9"),
			wantPass: false,
			wantIssues: 1,
		},
		{
			name: "multiple tracks, some with composer in title",
			actual: func() *domain.Album {
				composer, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("Vienna Phil", domain.RoleEnsemble)
				artists := []domain.Artist{composer, ensemble}
				
				track1, _ := domain.NewTrack(1, 1, "Symphony No. 1", artists)
				track2, _ := domain.NewTrack(1, 2, "Brahms: Symphony No. 2", artists)
				track3, _ := domain.NewTrack(1, 3, "Symphony No. 3", artists)
				
				album, _ := domain.NewAlbum("Brahms Symphonies", 1963)
				album.AddTrack(track1)
				album.AddTrack(track2)
				album.AddTrack(track3)
				return album
			}(),
			wantPass: false,
			wantIssues: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ComposerNotInTitle(tt.actual, tt.actual)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass && len(result.Issues()) != tt.wantIssues {
				t.Errorf("Issues = %d, want %d", len(result.Issues()), tt.wantIssues)
				for _, issue := range result.Issues() {
					t.Logf("  Issue: %s", issue.Message())
				}
			}
		})
	}
}

// Helper function to build an album with specific composer and title
func buildAlbumWithTitle(composerName, trackTitle string) *domain.Album {
	composer, _ := domain.NewArtist(composerName, domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Vienna Philharmonic", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Herbert von Karajan", domain.RoleConductor)
	
	artists := []domain.Artist{composer, ensemble, conductor}
	track, _ := domain.NewTrack(1, 1, trackTitle, artists)
	album, _ := domain.NewAlbum("Classical Album", 1963)
	album.AddTrack(track)
	return album
}

func TestExtractLastNames(t *testing.T) {
	tests := []struct {
		name         string
		composerName string
		want         []string
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
		t.Run(tt.name, func(t *testing.T) {
			got := extractLastNames(tt.composerName)
			if len(got) != len(tt.want) {
				t.Errorf("extractLastNames() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractLastNames()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
