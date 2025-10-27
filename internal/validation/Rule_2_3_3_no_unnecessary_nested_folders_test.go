package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoUnnecessaryNestedFolders(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name: "valid - no folders",
			Actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"02 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - single disc folder",
			Actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD1/02 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - multiple disc folders",
			Actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD2/01 - Track.flac",
				"CD3/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - Disc folder naming",
			Actual: buildAlbumWithFilenames(
				"Disc1/01 - Track.flac",
				"Disc2/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "valid - Disk folder naming",
			Actual: buildAlbumWithFilenames(
				"Disk1/01 - Track.flac",
				"Disk2/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "invalid - artist/album nesting",
			Actual: buildAlbumWithFilenames(
				"Beethoven/Symphonies/01 - Track.flac",
				"Beethoven/Symphonies/02 - Track.flac",
			),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name: "invalid - extra nested folder",
			Actual: buildAlbumWithFilenames(
				"Extra/CD1/01 - Track.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "invalid - year folder",
			Actual: buildAlbumWithFilenames(
				"1963/01 - Track.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name: "valid - DVD folder",
			Actual: buildAlbumWithFilenames(
				"DVD1/01 - Track.flac",
			),
			WantPass: true,
		},
		{
			Name: "invalid - some tracks with nesting",
			Actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"Artist/02 - Track.flac",
			),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.NoUnnecessaryNestedFolders(tt.Actual, tt.Actual)

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

func TestIsDiscFolder(t *testing.T) {
	tests := []struct {
		Name       string
		FolderName string
		Want       bool
	}{
		{"CD1", "CD1", true},
		{"CD2", "CD2", true},
		{"cd1", "cd1", true},
		{"Disc1", "Disc1", true},
		{"Disc2", "Disc2", true},
		{"disc1", "disc1", true},
		{"Disk1", "Disk1", true},
		{"DVD1", "DVD1", true},
		{"CD", "CD", true},
		{"Artist", "Artist", false},
		{"Album", "Album", false},
		{"1963", "1963", false},
		{"Beethoven", "Beethoven", false},
		{"CDextra", "CDextra", false}, // 'extra' contains non-digits
		{"cd10", "cd10", true},
		{"Disc", "Disc", true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := isDiscFolder(tt.FolderName)
			if got != tt.Want {
				t.Errorf("isDiscFolder(%q) = %v, want %v", tt.FolderName, got, tt.Want)
			}
		})
	}
}
