package domain

import (
	"strings"
	"testing"
)

func TestNewTrack(t *testing.T) {
	composer, _ := NewArtist("Felix Mendelssohn Bartholdy", RoleComposer)
	conductor, _ := NewArtist("Hans-Christoph Rademann", RoleConductor)
	ensemble, _ := NewArtist("RIAS Kammerchor Berlin", RoleEnsemble)

	tests := []struct {
		name    string
		disc    int
		track   int
		title   string
		artists []Artist
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid track with single composer",
			disc:    1,
			track:   1,
			title:   "Frohlocket, ihr Völker auf Erden, op.79/1",
			artists: []Artist{composer, ensemble, conductor},
			wantErr: false,
		},
		{
			name:    "multiple composers allowed (validated later)",
			disc:    1,
			track:   2,
			title:   "Some Work",
			artists: []Artist{composer, composer},
			wantErr: false,
		},
		{
			name:    "no composer allowed (validated later)",
			disc:    1,
			track:   3,
			title:   "Some Work",
			artists: []Artist{ensemble, conductor},
			wantErr: false,
		},
		{
			name:    "empty title allowed (validated later)",
			disc:    1,
			track:   4,
			title:   "",
			artists: []Artist{composer},
			wantErr: false,
		},
		{
			name:    "disc number zero allowed",
			disc:    0,
			track:   1,
			title:   "Some Work",
			artists: []Artist{composer},
			wantErr: false,
		},
		{
			name:    "track number zero allowed",
			disc:    1,
			track:   0,
			title:   "Some Work",
			artists: []Artist{composer},
			wantErr: false,
		},
		{
			name:    "no artists allowed (validated later)",
			disc:    1,
			track:   1,
			title:   "Some Work",
			artists: []Artist{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTrack(tt.disc, tt.track, tt.title, tt.artists)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTrack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewTrack() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				if got.Disc() != tt.disc {
					t.Errorf("NewTrack().Disc() = %v, want %v", got.Disc(), tt.disc)
				}
				if got.Track() != tt.track {
					t.Errorf("NewTrack().Track() = %v, want %v", got.Track(), tt.track)
				}
				if got.Title() != tt.title {
					t.Errorf("NewTrack().Title() = %v, want %v", got.Title(), tt.title)
				}
			}
		})
	}
}

func TestTrack_Composer(t *testing.T) {
	composer, _ := NewArtist("Johannes Brahms", RoleComposer)
	ensemble, _ := NewArtist("Berlin Philharmonic", RoleEnsemble)

	track, _ := NewTrack(1, 1, "Symphony No. 1, Op. 68", []Artist{composer, ensemble})

	got := track.Composer()
	if got.Name() != "Johannes Brahms" {
		t.Errorf("Track.Composer().Name() = %v, want %v", got.Name(), "Johannes Brahms")
	}
	if got.Role() != RoleComposer {
		t.Errorf("Track.Composer().Role() = %v, want %v", got.Role(), RoleComposer)
	}
}

func TestTrack_WithName(t *testing.T) {
	composer, _ := NewArtist("Anton Bruckner", RoleComposer)
	track, _ := NewTrack(1, 1, "Ave Maria", []Artist{composer})

	track = track.WithName("01 Ave Maria.flac")

	if got := track.Name(); got != "01 Ave Maria.flac" {
		t.Errorf("Track.Name() = %v, want %v", got, "01 Ave Maria.flac")
	}
}

func TestTrack_Validate_ComposerInTitle(t *testing.T) {
	composer, _ := NewArtist("Johann Sebastian Bach", RoleComposer)

	tests := []struct {
		name         string
		title        string
		wantErrorMsg string
	}{
		{
			name:         "composer last name in title",
			title:        "Bach: Goldberg Variations",
			wantErrorMsg: "Composer name",
		},
		{
			name:         "clean title without composer",
			title:        "Goldberg Variations, BWV 988",
			wantErrorMsg: "",
		},
		{
			name:         "composer last name mid-title",
			title:        "The Bach Variations",
			wantErrorMsg: "Composer name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			track, _ := NewTrack(1, 1, tt.title, []Artist{composer})
			issues := track.Validate()

			foundError := false
			for _, issue := range issues {
				if strings.Contains(issue.Message(), tt.wantErrorMsg) {
					foundError = true
					if issue.Level() != LevelError {
						t.Errorf("Composer in title should be ERROR level, got %v", issue.Level())
					}
					break
				}
			}

			if tt.wantErrorMsg != "" && !foundError {
				t.Errorf("Expected validation error containing %q, got no matching error", tt.wantErrorMsg)
			}
			if tt.wantErrorMsg == "" && foundError {
				t.Errorf("Expected no composer-in-title error, but got one")
			}
		})
	}
}

func TestTrack_Validate_FilenameTooLong(t *testing.T) {
	composer, _ := NewArtist("Dmitri Shostakovich", RoleComposer)
	track, _ := NewTrack(1, 1, "String Quartet No. 15", []Artist{composer})

	// Create a filename that exceeds 180 characters
	longName := strings.Repeat("a", 181) + ".flac"
	track = track.WithName(longName)

	issues := track.Validate()

	foundLengthError := false
	for _, issue := range issues {
		if strings.Contains(issue.Message(), "180 characters") {
			foundLengthError = true
			if issue.Level() != LevelError {
				t.Errorf("Filename length error should be ERROR level, got %v", issue.Level())
			}
		}
	}

	if !foundLengthError {
		t.Errorf("Expected filename length validation error")
	}
}
