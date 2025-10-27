package domain

import (
	"strings"
	"testing"
)

func TestNewTrack(t *testing.T) {
	composer := Artist{Name: "Felix Mendelssohn Bartholdy", Role: RoleComposer}
	conductor := Artist{Name: "Hans-Christoph Rademann", Role: RoleConductor}
	ensemble := Artist{Name: "RIAS Kammerchor Berlin", Role: RoleEnsemble}

	tests := []struct {
		Name    string
		Track   Track
	}{
		{
			Name:    "valid track with single composer",
			Track:   Track{Disc: 1, Track: 1, Title: "Frohlocket, ihr Völker auf Erden, op.79/1", Artists: []Artist{composer, ensemble, conductor}},
		},
		{
			Name:    "multiple composers allowed (validated later)",
			Track:   Track{Disc: 1, Track: 2, Title: "Some Work", Artists: []Artist{composer, composer}},
		},
		{
			Name:    "no composer allowed (validated later)",
			Track:   Track{Disc: 1, Track: 3, Title: "Some Work", Artists: []Artist{ensemble, conductor}},
		},
		{
			Name:    "empty title allowed (validated later)",
			Track:   Track{Disc: 1, Track: 4, Title: "", Artists: []Artist{composer}},
		},
		{
			Name:    "disc number zero allowed",
			Track:   Track{Disc: 0, Track: 1, Title: "Some Work", Artists: []Artist{composer}},
		},
		{
			Name:    "track number zero allowed",
			Track:   Track{Disc: 1, Track: 0, Title: "Some Work", Artists: []Artist{composer}},
		},
		{
			Name:    "no artists allowed (validated later)",
			Track:   Track{Disc: 1, Track: 1, Title: "Some Work", Artists: []Artist{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := tt.Track.Validate()
			if len(issues) > 0 {
				t.Errorf("%s: Track.Validate() returned %d issues, want none", tt.Name, len(issues))
			}
		})
	}
}

func TestTrack_Composer(t *testing.T) {
	track := Track{
		Disc: 1, 
		Track: 1, 
		Title: "Symphony No. 1, Op. 68", 
		Artists: []Artist{
			{Name: "Johannes Brahms", Role: RoleComposer}, 
			{Name: "Berlin Philharmonic", Role: RoleEnsemble},
		},
	}

	composers := track.Composers()
	if len(composers) != 1 {
		t.Errorf("Expected 1 composer, got %d", len(composers))
	}
	composer := composers[0]
	if composer.Name != "Johannes Brahms" {
		t.Errorf("Track.Composer().Name = %v, want %v", composer.Name, "Johannes Brahms")
	}
	if composer.Role != RoleComposer {
		t.Errorf("Track.Composer().Role() = %v, want %v", composer.Role, RoleComposer)
	}
}

func TestTrack_Validate_ComposerInTitle(t *testing.T) {
	tests := []struct {
		Name         string
		Title        string
		WantErrorMsg string
	}{
		{
			Name:         "composer last name in title",
			Title:        "Bach: Goldberg Variations",
			WantErrorMsg: "Composer name",
		},
		{
			Name:         "clean title without composer",
			Title:        "Goldberg Variations, BWV 988",
			WantErrorMsg: "",
		},
		{
			Name:         "composer last name mid-title",
			Title:        "The Bach Variations",
			WantErrorMsg: "Composer name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			track := Track{
				Disc: 1, 
				Track: 1, 
				Title: tt.Title, 
				Artists: []Artist{{Name: "Johann Sebastian Bach", Role: RoleComposer}},
			}
			issues := track.Validate()

			foundError := false
			messages := []string{}
			for _, issue := range issues {
				messages = append(messages, issue.Message)
				if strings.Contains(issue.Message, tt.WantErrorMsg) {
					foundError = true
					if issue.Level != LevelError {
						t.Errorf("Composer in title should be ERROR level, got %v", issue.Level)
					}
					break
				}
			}

			if tt.WantErrorMsg != "" && !foundError {
				t.Errorf("Expected validation error containing %q, got no matching error in %v", tt.WantErrorMsg, messages)
			}
			if tt.WantErrorMsg == "" && foundError {
				t.Errorf("Expected no composer-in-title error, but got one in %v", messages)
			}
		})
	}
}

func TestTrack_Validate_FilenameTooLong(t *testing.T) {
	track := Track{
		Disc: 1, 
		Track: 1, 
		Title: "String Quartet No. 15", 
		Artists: []Artist{{Name: "Dmitri Shostakovich", Role: RoleComposer}},
		Name: strings.Repeat("a", 181) + ".flac",
	}

	issues := track.Validate()

	foundLengthError := false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "180 characters") {
			foundLengthError = true
			if issue.Level != LevelError {
				t.Errorf("Filename length error should be ERROR level, got %v", issue.Level)
			}
		}
	}

	if !foundLengthError {
		t.Errorf("Expected filename length validation error")
	}
}
