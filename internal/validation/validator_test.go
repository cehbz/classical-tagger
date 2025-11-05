package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		Name           string
		SetupTorrent *domain.Torrent
		WantErrorCount int
		WantWarnCount  int
	}{
		{
			Name: "valid complete album",
			SetupTorrent: &domain.Torrent{
				Title:        "Test Album",
				OriginalYear: 2013,
				Edition: &domain.Edition{
					Label:         "test label",
					Year:          2013,
					CatalogNumber: "HMC902170",
				},
				Files: []domain.FileLike{
					&domain.Track{
						File: domain.File{Path: "01 Frohlocket, Op. 79-1.flac"},
						Disc:  1,
						Track: 1,
						Title: "Frohlocket, Op. 79/1",
						Artists: []domain.Artist{
							{Name: "Felix Mendelssohn", Role: domain.RoleComposer},
							{Name: "RIAS Kammerchor", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantErrorCount: 0, // No errors - RIAS is now recognized as acronym
			WantWarnCount:  3,
		},
		{
			Name: "missing edition",
			SetupTorrent: &domain.Torrent{
				Title:        "Test Album",
				OriginalYear: 2013,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Symphony No. 1",
						Artists: []domain.Artist{
							{Name: "Johannes Brahms", Role: domain.RoleComposer},
						},
					},
				},
			},
			WantErrorCount: 2,
			WantWarnCount:  5,
		},
		{
			Name: "composer in title",
			SetupTorrent: &domain.Torrent{
				Title:        "Test Album",
				OriginalYear: 2013,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Bach: Goldberg Variations",
						Artists: []domain.Artist{
							{Name: "Johann Sebastian Bach", Role: domain.RoleComposer},
						},
					},
				},
			},
			WantErrorCount: 3,
			WantWarnCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			torrent := tt.SetupTorrent
			issues := Check(torrent, nil)

			errorCount := 0
			warnCount := 0
			for _, issue := range issues {
				switch issue.Level {
case domain.LevelError:
					errorCount++
				case domain.LevelWarning:
					warnCount++
				}
			}

			if errorCount != tt.WantErrorCount {
				t.Errorf("ValidateMetadata() error count = %d, want %d", errorCount, tt.WantErrorCount)
			}
			if warnCount != tt.WantWarnCount {
				t.Errorf("ValidateMetadata() warning count = %d, want %d", warnCount, tt.WantWarnCount)
			}
		})
	}
}
