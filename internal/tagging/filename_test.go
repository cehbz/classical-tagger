package tagging

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		Name        string
		Track       *domain.Track
		TotalTracks int
		Want        string
	}{
		{
			Name: "single digit track number, <=9 tracks",
			Track: &domain.Track{
				Track: 1,
				Title: "Symphony No. 5",
			},
			TotalTracks: 5,
			Want:        "1 - Symphony No. 5.flac",
		},
		{
			Name: "single digit track number, >9 tracks (leading zero)",
			Track: &domain.Track{
				Track: 1,
				Title: "Symphony No. 5",
			},
			TotalTracks: 15,
			Want:        "01 - Symphony No. 5.flac",
		},
		{
			Name: "double digit track number, >9 tracks",
			Track: &domain.Track{
				Track: 12,
				Title: "Finale",
			},
			TotalTracks: 15,
			Want:        "12 - Finale.flac",
		},
		{
			Name: "title with invalid characters",
			Track: &domain.Track{
				Track: 3,
				Title: "Track: \"Special\" / Path\\Name",
			},
			TotalTracks: 10,
			Want:        "03 - Track Special PathName.flac",
		},
		{
			Name: "empty title",
			Track: &domain.Track{
				Track: 1,
				Title: "",
			},
			TotalTracks: 5,
			Want:        "1 - Untitled.flac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := GenerateFilename(tt.Track, tt.TotalTracks)
			if got != tt.Want {
				t.Errorf("GenerateFilename() = %q, want %q", got, tt.Want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "normal title",
			Input: "Symphony No. 5",
			Want:  "Symphony No. 5",
		},
		{
			Name:  "invalid characters",
			Input: "Track: \"Name\" / Path\\File",
			Want:  "Track Name PathFile",
		},
		{
			Name:  "leading/trailing spaces and dots",
			Input: "  .  Title  .  ",
			Want:  "Title",
		},
		{
			Name:  "multiple spaces",
			Input: "Multiple    Spaces   Here",
			Want:  "Multiple Spaces Here",
		},
		{
			Name:  "Windows reserved name",
			Input: "CON",
			Want:  "_CON",
		},
		{
			Name:  "empty string",
			Input: "",
			Want:  "",
		},
		{
			Name:  "very long title",
			Input: "This is a very long title that exceeds the maximum length limit and should be truncated appropriately to fit within reasonable filename constraints",
			Want:  "This is a very long title that exceeds the maximum length limit and should be truncated appropriately to fit within reasonable filename constraints",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := SanitizeFilename(tt.Input)
			if got != tt.Want {
				t.Errorf("SanitizeFilename() = %q, want %q", got, tt.Want)
			}
		})
	}
}

func TestGenerateDiscSubdirectoryName(t *testing.T) {
	tests := []struct {
		Name      string
		DiscNum   int
		DiscTitle string
		Want      string
	}{
		{
			Name:      "disc number only",
			DiscNum:   1,
			DiscTitle: "",
			Want:      "Disc 1",
		},
		{
			Name:      "disc number with title",
			DiscNum:   2,
			DiscTitle: "Live Performance",
			Want:      "Live Performance",
		},
		{
			Name:      "disc title with invalid characters",
			DiscNum:   3,
			DiscTitle: "Disc: \"Special\" / Name",
			Want:      "Disc Special Name",
		},
		{
			Name:      "empty disc title falls back to number",
			DiscNum:   1,
			DiscTitle: "   ",
			Want:      "Disc 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := GenerateDiscSubdirectoryName(tt.DiscNum, tt.DiscTitle)
			if got != tt.Want {
				t.Errorf("GenerateDiscSubdirectoryName() = %q, want %q", got, tt.Want)
			}
		})
	}
}
