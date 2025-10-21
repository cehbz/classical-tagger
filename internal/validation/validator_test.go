package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestAlbumValidator_ValidateMetadata(t *testing.T) {
	validator := NewAlbumValidator()
	
	tests := []struct {
		name           string
		setupAlbum     func() *domain.Album
		wantErrorCount int
		wantWarnCount  int
	}{
		{
			name: "valid complete album",
			setupAlbum: func() *domain.Album {
				album, _ := domain.NewAlbum("Test Album", 2013)
				edition, _ := domain.NewEdition("harmonia mundi", 2013)
				edition = edition.WithCatalogNumber("HMC902170")
				album = album.WithEdition(edition)
				
				composer, _ := domain.NewArtist("Felix Mendelssohn", domain.RoleComposer)
				ensemble, _ := domain.NewArtist("RIAS Kammerchor", domain.RoleEnsemble)
				track, _ := domain.NewTrack(1, 1, "Frohlocket, Op. 79/1", []domain.Artist{composer, ensemble})
				track = track.WithName("01 Frohlocket, Op. 79-1.flac")
				album.AddTrack(track)
				
				return album
			},
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name: "missing edition",
			setupAlbum: func() *domain.Album {
				album, _ := domain.NewAlbum("Test Album", 2013)
				composer, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Symphony No. 1", []domain.Artist{composer})
				album.AddTrack(track)
				return album
			},
			wantErrorCount: 0,
			wantWarnCount:  1, // missing edition
		},
		{
			name: "composer in title",
			setupAlbum: func() *domain.Album {
				album, _ := domain.NewAlbum("Test Album", 2013)
				composer, _ := domain.NewArtist("Johann Sebastian Bach", domain.RoleComposer)
				track, _ := domain.NewTrack(1, 1, "Bach: Goldberg Variations", []domain.Artist{composer})
				album.AddTrack(track)
				return album
			},
			wantErrorCount: 1, // composer in title
			wantWarnCount:  1, // missing edition
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			album := tt.setupAlbum()
			issues := validator.ValidateMetadata(album)
			
			errorCount := 0
			warnCount := 0
			for _, issue := range issues {
				if issue.Level() == domain.LevelError {
					errorCount++
				} else if issue.Level() == domain.LevelWarning {
					warnCount++
				}
			}
			
			if errorCount != tt.wantErrorCount {
				t.Errorf("ValidateMetadata() error count = %d, want %d", errorCount, tt.wantErrorCount)
			}
			if warnCount != tt.wantWarnCount {
				t.Errorf("ValidateMetadata() warning count = %d, want %d", warnCount, tt.wantWarnCount)
			}
		})
	}
}
