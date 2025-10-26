package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoUnnecessaryNestedFolders(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - no folders",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"02 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - single disc folder",
			actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD1/02 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - multiple disc folders",
			actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD2/01 - Track.flac",
				"CD3/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - Disc folder naming",
			actual: buildAlbumWithFilenames(
				"Disc1/01 - Track.flac",
				"Disc2/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "valid - Disk folder naming",
			actual: buildAlbumWithFilenames(
				"Disk1/01 - Track.flac",
				"Disk2/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "invalid - artist/album nesting",
			actual: buildAlbumWithFilenames(
				"Beethoven/Symphonies/01 - Track.flac",
				"Beethoven/Symphonies/02 - Track.flac",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "invalid - extra nested folder",
			actual: buildAlbumWithFilenames(
				"Extra/CD1/01 - Track.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - year folder",
			actual: buildAlbumWithFilenames(
				"1963/01 - Track.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "valid - DVD folder",
			actual: buildAlbumWithFilenames(
				"DVD1/01 - Track.flac",
			),
			wantPass: true,
		},
		{
			name: "invalid - some tracks with nesting",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"Artist/02 - Track.flac",
			),
			wantPass:   false,
			wantIssues: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.NoUnnecessaryNestedFolders(tt.actual, tt.actual)
			
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

func TestIsDiscFolder(t *testing.T) {
	tests := []struct {
		name       string
		folderName string
		want       bool
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
		t.Run(tt.name, func(t *testing.T) {
			got := isDiscFolder(tt.folderName)
			if got != tt.want {
				t.Errorf("isDiscFolder(%q) = %v, want %v", tt.folderName, got, tt.want)
			}
		})
	}
}
