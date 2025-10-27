package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_FilenamesMatchTitles(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - exact match",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Symphony No. 5.flac"},
					{"Concerto in D", "02 - Concerto in D.flac"},
				},
			),
			WantPass: true,
		},
		{
			Name: "valid - minor punctuation differences",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5 in C Minor", "01 - Symphony No 5 in C Minor.flac"},
					{"Concerto: Allegro", "02 - Concerto Allegro.flac"},
				},
			),
			WantPass: true,
		},
		{
			Name: "valid - filename is abbreviation of title",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5 in C Minor, Op. 67", "01 - Symphony No. 5.flac"},
				},
			),
			WantPass: true,
		},
		{
			Name: "valid - minor typo (edit distance ≤ 3)",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Sympony No. 5.flac"}, // 1 char difference
				},
			),
			WantPass: true,
		},
		{
			Name: "invalid - completely different title",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Concerto in D.flac"},
					{"Concerto in D", "02 - Symphony No. 9.flac"},
				},
			),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name: "invalid - major misspelling",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - Symphonieee No. 5.flac"}, // Too many differences
				},
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "valid - case differences",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01 - SYMPHONY NO. 5.flac"},
				},
			),
			WantPass: true,
		},
		{
			Name: "valid - different separators",
			Actual: buildAlbumWithTitlesAndFilenames(
				[]titleFile{
					{"Symphony No. 5", "01. Symphony No. 5.flac"},
					{"Concerto in D", "02_Concerto in D.flac"},
				},
			),
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.FilenamesMatchTitles(tt.Actual, tt.Actual)

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

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		Input string
		Want  string
	}{
		{"Symphony No. 5", "symphony no 5"},
		{"Concerto: Allegro!", "concerto allegro"},
		{"Work (with parentheses)", "work with parentheses"},
		{"Title, with commas", "title with commas"},
		{"Multiple   spaces", "multiple spaces"},
		{"Title's Possessive", "titles possessive"},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			got := normalizeTitle(tt.Input)
			if got != tt.Want {
				t.Errorf("normalizeTitle(%q) = %q, want %q", tt.Input, got, tt.Want)
			}
		})
	}
}

func TestTitlesMatch(t *testing.T) {
	tests := []struct {
		Name   string
		Title1 string
		Title2 string
		Want   bool
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
		t.Run(tt.Name, func(t *testing.T) {
			got := titlesMatch(tt.Title1, tt.Title2)
			if got != tt.Want {
				t.Errorf("titlesMatch(%q, %q) = %v, want %v", tt.Title1, tt.Title2, got, tt.Want)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		S1   string
		S2   string
		Want int
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
		t.Run(tt.S1+"_"+tt.S2, func(t *testing.T) {
			got := levenshteinDistance(tt.S1, tt.S2)
			if got != tt.Want {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.S1, tt.S2, got, tt.Want)
			}
		})
	}
}

// titleFile pairs track title with filename
type titleFile struct {
	Title    string
	Filename string
}

// buildAlbumWithTitlesAndFilenames creates album with specific title/filename pairs
func buildAlbumWithTitlesAndFilenames(titleFiles []titleFile) *domain.Album {
	tracks := make([]*domain.Track, len(titleFiles))
	for i, tf := range titleFiles {
		tracks[i] = &domain.Track{
			Disc:  1,
			Track: i + 1,
			Title: tf.Title,
			Artists: []domain.Artist{
				domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
				domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
			},
		}
	}

	return &domain.Album{
		Title:        "Test Album",
		OriginalYear: 1963,
		Tracks:       tracks,
	}
}
