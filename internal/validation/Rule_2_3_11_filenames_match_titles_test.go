package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenamesMatchTitles(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - exact match",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Symphony No. 5.flac"},
					{"Concerto in D", "02 - Concerto in D.flac"},
				},
			),
			wantPass: true,
		},
		{
			name: "valid - minor punctuation differences",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5 in C Minor", "01 - Symphony No 5 in C Minor.flac"},
					{"Concerto: Allegro", "02 - Concerto Allegro.flac"},
				},
			),
			wantPass: true,
		},
		{
			name: "valid - filename is abbreviation of title",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5 in C Minor, Op. 67", "01 - Symphony No. 5.flac"},
				},
			),
			wantPass: true,
		},
		{
			name: "valid - minor typo (edit distance ≤ 3)",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Sympony No. 5.flac"}, // 1 char difference
				},
			),
			wantPass: true,
		},
		{
			name: "invalid - completely different title",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Concerto in D.flac"},
					{"Concerto in D", "02 - Symphony No. 9.flac"},
				},
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "invalid - major misspelling",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Symphonieee No. 5.flac"}, // Too many differences
				},
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "valid - case differences",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - SYMPHONY NO. 5.flac"},
				},
			),
			wantPass: true,
		},
		{
			name: "valid - different separators",
			actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01. Symphony No. 5.flac"},
					{"Concerto in D", "02_Concerto in D.flac"},
				},
			),
			wantPass: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.FilenamesMatchTitles(tt.actual, tt.actual)
			
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

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Symphony No. 5", "symphony no 5"},
		{"Concerto: Allegro!", "concerto allegro"},
		{"Work (with parentheses)", "work with parentheses"},
		{"Title, with commas", "title with commas"},
		{"Multiple   spaces", "multiple spaces"},
		{"Title's Possessive", "titles possessive"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeTitle(tt.input)
			if got != tt.want {
				t.Errorf("normalizeTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTitlesMatch(t *testing.T) {
	tests := []struct {
		name   string
		title1 string
		title2 string
		want   bool
	}{
		{"exact match", "symphony no 5", "symphony no 5", true},
		{"one is substring", "symphony no 5", "symphony no 5 in c minor", true},
		{"substring reverse", "symphony no 5 in c minor", "symphony no 5", true},
		{"edit distance 1", "symphony", "symphoy", true},
		{"edit distance 3", "symphony", "symphny", true},
		{"edit distance >3", "symphony", "symph", false},
		{"completely different", "symphony", "concerto", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titlesMatch(tt.title1, tt.title2)
			if got != tt.want {
				t.Errorf("titlesMatch(%q, %q) = %v, want %v", tt.title1, tt.title2, got, tt.want)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1   string
		s2   string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "adc", 1},
		{"symphony", "symphoy", 1},
		{"symphony", "symphny", 1},
		{"kitten", "sitting", 3},
	}
	
	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			got := levenshteinDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

// titleFile pairs track title with filename
type titleFile struct {
	title    string
	filename string
}

// buildAlbumWithTitlesAndFilenames creates album with specific title/filename pairs
func buildAlbumWithTitlesAndFilenames(titleFiles []titleFile) *domain.Album {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}
	
	album, _ := domain.NewAlbum("Test Album", 1963)
	for i, tf := range titleFiles {
		track, _ := domain.NewTrack(1, i+1, tf.title, artists)
		track = track.WithName(tf.filename)
		album.AddTrack(track)
	}
	
	return album
}
