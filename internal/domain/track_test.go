package domain

import (
	"testing"
)

func TestNewTrack(t *testing.T) {
	composer := Artist{Name: "Felix Mendelssohn Bartholdy", Role: RoleComposer}
	conductor := Artist{Name: "Hans-Christoph Rademann", Role: RoleConductor}
	ensemble := Artist{Name: "RIAS Kammerchor Berlin", Role: RoleEnsemble}

	tests := []struct {
		Name  string
		Track Track
	}{
		{
			Name:  "valid track with single composer",
			Track: Track{Disc: 1, Track: 1, Title: "Frohlocket, ihr Völker auf Erden, op.79/1", Artists: []Artist{composer, ensemble, conductor}},
		},
		{
			Name:  "multiple composers allowed (validated later)",
			Track: Track{Disc: 1, Track: 2, Title: "Some Work", Artists: []Artist{composer, composer}},
		},
		{
			Name:  "no composer allowed (validated later)",
			Track: Track{Disc: 1, Track: 3, Title: "Some Work", Artists: []Artist{ensemble, conductor}},
		},
		{
			Name:  "empty title allowed (validated later)",
			Track: Track{Disc: 1, Track: 4, Title: "", Artists: []Artist{composer}},
		},
		{
			Name:  "disc number zero allowed",
			Track: Track{Disc: 0, Track: 1, Title: "Some Work", Artists: []Artist{composer}},
		},
		{
			Name:  "track number zero allowed",
			Track: Track{Disc: 1, Track: 0, Title: "Some Work", Artists: []Artist{composer}},
		},
		{
			Name:  "no artists allowed (validated later)",
			Track: Track{Disc: 1, Track: 1, Title: "Some Work", Artists: []Artist{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Track construction is simple - no errors
			// Validation happens later via validation package
			_ = tt.Track
		})
	}
}

func TestTrack_Composer(t *testing.T) {
	track := Track{
		Disc:  1,
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
