package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_NoArchiveFiles(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name       string
		actual     *domain.Album
		wantPass   bool
		wantIssues int
	}{
		{
			name: "valid - only audio files",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"02 - Track.mp3",
				"03 - Track.wav",
			),
			wantPass: true,
		},
		{
			name: "invalid - zip file present",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"booklet.zip",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - rar file present",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"artwork.rar",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - 7z file present",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"scans.7z",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - tar.gz file present",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"files.tar.gz",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - multiple archive files",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"artwork.zip",
				"booklet.rar",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "valid - .log and .cue files are OK",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"Album.log",
				"Album.cue",
			),
			wantPass: true,
		},
		{
			name: "valid - image files are OK",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"cover.jpg",
				"back.png",
			),
			wantPass: true,
		},
		{
			name: "invalid - case insensitive check",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"files.ZIP",
				"data.RAR",
			),
			wantPass:   false,
			wantIssues: 2,
		},
		{
			name: "invalid - archive in subdirectory",
			actual: buildAlbumWithFilenames(
				"CD1/01 - Track.flac",
				"CD1/booklet.zip",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - tgz shorthand",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"files.tgz",
			),
			wantPass:   false,
			wantIssues: 1,
		},
		{
			name: "invalid - bz2 compression",
			actual: buildAlbumWithFilenames(
				"01 - Track.flac",
				"files.tar.bz2",
			),
			wantPass:   false,
			wantIssues: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.NoArchiveFiles(tt.actual, tt.actual)
			
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
