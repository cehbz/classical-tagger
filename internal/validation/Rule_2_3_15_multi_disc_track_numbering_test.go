package validation

import (
	"fmt"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_MultiDiscTrackNumbering(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		name         string
		actual       *domain.Album
		wantPass     bool
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid - single disc",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {1, 3}},
			),
			wantPass: true,
		},
		{
			name: "valid - two discs, both start at 1",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {2, 1}, {2, 2}},
			),
			wantPass: true,
		},
		{
			name: "valid - three discs, all start at 1",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{
					{1, 1}, {1, 2},
					{2, 1}, {2, 2}, {2, 3},
					{3, 1}, {3, 2},
				},
			),
			wantPass: true,
		},
		{
			name: "invalid - disc 2 doesn't start at 1",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {2, 3}, {2, 4}},
			),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "invalid - disc 1 doesn't start at 1",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 2}, {1, 3}, {2, 1}, {2, 2}},
			),
			wantPass:   false,
			wantErrors: 1,
		},
		{
			name: "invalid - both discs don't start at 1",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 5}, {1, 6}, {2, 10}, {2, 11}},
			),
			wantPass:   false,
			wantErrors: 2,
		},
		{
			name: "warning - gap in track numbering",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {1, 4}}, // Missing track 3
			),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name: "invalid - missing disc in sequence",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{{1, 1}, {1, 2}, {3, 1}, {3, 2}}, // No disc 2
			),
			wantPass:   false,
			wantErrors: 1, // Missing disc 2
		},
		{
			name: "valid - large multi-disc set",
			actual: buildAlbumWithDiscTracks(
				[]discTrack{
					{1, 1}, {1, 2}, {1, 3},
					{2, 1}, {2, 2},
					{3, 1}, {3, 2}, {3, 3}, {3, 4},
					{4, 1}, {4, 2},
				},
			),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.MultiDiscTrackNumbering(tt.actual, tt.actual)

			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}

			if !tt.wantPass {
				errorCount := 0
				warningCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelError {
						errorCount++
					} else if issue.Level() == domain.LevelWarning {
						warningCount++
					}
				}

				if tt.wantErrors > 0 && errorCount != tt.wantErrors {
					t.Errorf("Errors = %d, want %d", errorCount, tt.wantErrors)
				}
				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}

				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

// discTrack represents a track with disc and track number
type discTrack struct {
	disc     int
	trackNum int
}

// buildAlbumWithDiscTracks creates an album with specific disc/track combinations
func buildAlbumWithDiscTracks(discTracks []discTrack) *domain.Album {
	tracks := make([]*domain.Track, len(discTracks))
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	artists := []domain.Artist{composer, ensemble}

	for i, dt := range discTracks {
		track, _ := domain.NewTrack(dt.disc, dt.trackNum, fmt.Sprintf("Track D%d-T%d", dt.disc, dt.trackNum), artists)
		track = track.WithName(fmt.Sprintf("CD%d/%02d - Track.flac", dt.disc, dt.trackNum))
		tracks[i] = track
	}

	album, _ := domain.NewAlbum("Multi-Disc Album", 1963)
	for _, track := range tracks {
		album.AddTrack(track)
	}
	return album
}
