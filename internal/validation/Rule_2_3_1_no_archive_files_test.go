package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// TestRules_NoArchiveFiles is disabled because archive files are not included in metadata.
// The metadata JSON only contains tracks (FLAC files), not all files in the directory.
// This rule requires filesystem access or expanded metadata structure.
// See TODO.md for details.
func TestRules_NoArchiveFiles(t *testing.T) {
	t.Skip("Skipping NoArchiveFiles test - archive files not in metadata. See TODO.md")
	rules := NewRules()

	tests := []struct {
		Name       string
		Actual     *domain.Album
		WantPass   bool
		WantIssues int
	}{
		{
			Name:       "valid - only audio files",
			Actual:     buildAlbumWithFilenames("01 - Track.flac"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "invalid - zip file present",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "booklet.zip"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - rar file present",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "artwork.rar"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - 7z file present",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "scans.7z"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - tar.gz file present",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "files.tar.gz"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - multiple archive files",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "artwork.zip", "booklet.rar"),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name:       "valid - .log and .cue files are OK",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "Album.log", "Album.cue"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "valid - image files are OK",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "cover.jpg", "back.png"),
			WantPass:   true,
			WantIssues: 0,
		},
		{
			Name:       "invalid - case insensitive check",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "files.ZIP", "data.RAR"),
			WantPass:   false,
			WantIssues: 2,
		},
		{
			Name:       "invalid - archive in subdirectory",
			Actual:     buildAlbumWithFilenames("CD1/01 - Track.flac", "CD1/booklet.zip"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - tgz shorthand",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "files.tgz"),
			WantPass:   false,
			WantIssues: 1,
		},
		{
			Name:       "invalid - bz2 compression",
			Actual:     buildAlbumWithFilenames("01 - Track.flac", "files.tar.bz2"),
			WantPass:   false,
			WantIssues: 1,
		},
	}

	for _, tt := range tests {
		for _, track := range tt.Actual.Tracks {
			t.Run(tt.Name, func(t *testing.T) {
				result := rules.NoArchiveFiles(track, nil, tt.Actual, nil)

				// NoArchiveFiles checks the entire album for archive files
				// So each track reports the same total number of archive files found
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
}
