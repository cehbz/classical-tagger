package scraping

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestLocalExtractor_ParseDirectoryName(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name           string
		DirPath        string
		WantFolderName string
		WantTitle      string
		WantYear       int
	}{
		{
			Name:           "standard format with brackets",
			DirPath:        "/music/Beethoven - Symphony No. 5 [1963] [FLAC]",
			WantFolderName: "Beethoven - Symphony No. 5 [1963] [FLAC]",
			WantTitle:      "Beethoven - Symphony No. 5",
			WantYear:       1963,
		},
		{
			Name:           "parentheses for year",
			DirPath:        "music/Bach - Goldberg Variations (1741)",
			WantFolderName: "Bach - Goldberg Variations (1741)",
			WantTitle:      "Bach - Goldberg Variations",
			WantYear:       1741,
		},
		{
			Name:           "no year",
			DirPath:        "files/classical/music/Mozart - Piano Concertos [FLAC]",
			WantFolderName: "Mozart - Piano Concertos [FLAC]",
			WantTitle:      "Mozart - Piano Concertos",
			WantYear:       0,
		},
		{
			Name:           "no format indicator",
			DirPath:        "./Vivaldi - Four Seasons [1989]",
			WantFolderName: "Vivaldi - Four Seasons [1989]",
			WantTitle:      "Vivaldi - Four Seasons",
			WantYear:       1989,
		},
		{
			Name:           "complex format with bit depth",
			DirPath:        "J.S. Bach - Brandenburg Concertos [1982] [FLAC] [24-96]",
			WantFolderName: "J.S. Bach - Brandenburg Concertos [1982] [FLAC] [24-96]",
			WantTitle:      "J.S. Bach - Brandenburg Concertos",
			WantYear:       1982,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotFolder, gotTitle, gotYear := extractor.parseDirectoryName(tt.DirPath)

			if gotFolder != tt.WantFolderName {
				t.Errorf("parseDirectoryName() folder = %v, want %v", gotFolder, tt.WantFolderName)
			}

			if gotTitle != tt.WantTitle {
				t.Errorf("parseDirectoryName() title = %v, want %v", gotTitle, tt.WantTitle)
			}
			if gotYear != tt.WantYear {
				t.Errorf("parseDirectoryName() year = %v, want %v", gotYear, tt.WantYear)
			}
		})
	}
}

func TestLocalExtractor_ExtractTrackNumberFromFilename(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name     string
		Filename string
		Want     int
	}{
		{"space separator", "/music/01 Track Title.flac", 1},
		{"dash separator", "/music/01-Track Title.flac", 1},
		{"dot separator", "/music/01.Track Title.flac", 1},
		{"underscore separator", "/music/01_Track Title.flac", 1},
		{"two digits", "/music/12 Track Title.flac", 12},
		{"three digits", "/music/123 Track Title.flac", 123},
		{"no number", "/music/Track Title.flac", 0},
		{"number in middle", "/music/Track 01 Title.flac", 0}, // Should not match
		{"padded", "/music/001 Track Title.flac", 1},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.extractTrackNumberFromFilename(tt.Filename)
			if got != tt.Want {
				t.Errorf("extractTrackNumberFromFilename(%q) = %v, want %v", tt.Filename, got, tt.Want)
			}
		})
	}
}

func TestLocalExtractor_ExtractDiscFromPath(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name string
		Path string
		Want int
	}{
		{"CD1", "/music/Album/CD1/01 Track.flac", 1},
		{"CD2", "/music/Album/CD2/01 Track.flac", 2},
		{"Disc 1", "/music/Album/Disc 1/01 Track.flac", 1},
		{"Disc 2", "/music/Album/Disc 2/01 Track.flac", 2},
		{"Disk 3", "/music/Album/Disk 3/01 Track.flac", 3},
		{"no disc", "/music/Album/01 Track.flac", 1}, // Default
		{"mixed case", "/music/Album/cd1/01 Track.flac", 1},
		{"with space", "/music/Album/CD 10/01 Track.flac", 10},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.extractDiscFromPath(tt.Path)
			if got != tt.Want {
				t.Errorf("extractDiscFromPath(%q) = %v, want %v", tt.Path, got, tt.Want)
			}
		})
	}
}

func TestLocalExtractor_ExtractTitleFromFilename(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name     string
		Filename string
		Want     string
	}{
		{"with track number", "/music/01 Symphony No. 5.flac", "Symphony No. 5"},
		{"with dash", "/music/01-Symphony No. 5.flac", "Symphony No. 5"},
		{"with dot", "/music/01.Symphony No. 5.flac", "Symphony No. 5"},
		{"no track number", "/music/Symphony No. 5.flac", "Symphony No. 5"},
		{"padded number", "/music/001 Concerto.flac", "Concerto"},
		{"with underscore", "/music/12_Piano Sonata.flac", "Piano Sonata"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.extractTitleFromFilename(tt.Filename)
			if got != tt.Want {
				t.Errorf("extractTitleFromFilename(%q) = %v, want %v", tt.Filename, got, tt.Want)
			}
		})
	}
}

func TestLocalExtractor_InferRoleFromName(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name       string
		ArtistName string
		Want       domain.Role
	}{
		{"conductor explicit", "Herbert von Karajan, conductor", domain.RoleConductor},
		{"orchestra", "Berlin Philharmonic Orchestra", domain.RoleEnsemble},
		{"philharmonic", "Vienna Philharmonic", domain.RoleEnsemble},
		{"symphony", "London Symphony Orchestra", domain.RoleEnsemble},
		{"choir", "RIAS Kammerchor", domain.RoleEnsemble},
		{"chorus", "Westminster Choir", domain.RoleEnsemble},
		{"quartet", "Emerson String Quartet", domain.RoleEnsemble},
		{"trio", "Beaux Arts Trio", domain.RoleEnsemble},
		{"chamber", "English Chamber Orchestra", domain.RoleEnsemble},
		{"consort", "Gabrieli Consort", domain.RoleEnsemble},
		{"soloist", "Maurizio Pollini", domain.RoleSoloist},
		{"soloist name", "Anne-Sophie Mutter", domain.RoleSoloist},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.inferRoleFromName(tt.ArtistName)
			if got != tt.Want {
				t.Errorf("inferRoleFromName(%q) = %v, want %v", tt.ArtistName, got, tt.Want)
			}
		})
	}
}

func TestLocalExtractor_ParseArtistField(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name      string
		Field     string
		WantCount int
		WantFirst string
		WantRole  domain.Role
	}{
        {
            Name:      "semicolon separated",
            Field:     "Pollini; Berlin Phil; Karajan",
            WantCount: 3,
            WantFirst: "Pollini",
            WantRole:  domain.RoleUnknown,
        },
        {
            Name:      "comma separated",
            Field:     "Pollini, Berlin Philharmonic, Karajan",
            WantCount: 3,
            WantFirst: "Pollini",
            WantRole:  domain.RoleUnknown,
        },
        {
            Name:      "single artist",
            Field:     "Maurizio Pollini",
            WantCount: 1,
            WantFirst: "Maurizio Pollini",
            WantRole:  domain.RoleUnknown,
        },
        {
            Name:      "with ensemble",
            Field:     "RIAS Kammerchor; Hans-Christoph Rademann",
            WantCount: 2,
            WantFirst: "RIAS Kammerchor",
            WantRole:  domain.RoleUnknown,
        },
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.parseArtistField(tt.Field)

			if len(got) != tt.WantCount {
				t.Errorf("parseArtistField(%q) returned %d artists, want %d", tt.Field, len(got), tt.WantCount)
			}

			if len(got) > 0 {
				if got[0].Name != tt.WantFirst {
					t.Errorf("First artist name = %v, want %v", got[0].Name, tt.WantFirst)
				}
				if got[0].Role != tt.WantRole {
					t.Errorf("First artist role = %v, want %v", got[0].Role, tt.WantRole)
				}
			}
		})
	}
}

func TestLocalExtractor_ExtractEditionFromComment(t *testing.T) {
	extractor := NewLocalExtractor()

	tests := []struct {
		Name        string
		Comment     string
		WantLabel   string
		WantCatalog string
		WantNil     bool
	}{
		{
			Name:        "label and catalog",
			Comment:     "Label: Deutsche Grammophon\nCatalog: 479 1234",
			WantLabel:   "Deutsche Grammophon",
			WantCatalog: "479 1234",
			WantNil:     false,
		},
		{
			Name:        "label only",
			Comment:     "Label: Harmonia Mundi",
			WantLabel:   "Harmonia Mundi",
			WantCatalog: "",
			WantNil:     false,
		},
		{
			Name:        "catalog only",
			Comment:     "Catalog Number: HMC902170",
			WantLabel:   "",
			WantCatalog: "HMC902170",
			WantNil:     false,
		},
		{
			Name:        "case insensitive",
			Comment:     "LABEL: Test Label\nCATALOG: ABC123",
			WantLabel:   "Test Label",
			WantCatalog: "ABC123",
			WantNil:     false,
		},
		{
			Name:    "no edition data",
			Comment: "Just some random comment",
			WantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractor.extractEditionFromComment(tt.Comment)

			if tt.WantNil {
				if got != nil {
					t.Error("extractEditionFromComment() should return nil, got non-nil")
				}
				return
			}

			if got == nil {
				t.Fatal("extractEditionFromComment() returned nil, want non-nil")
			}

			if got.Label != tt.WantLabel {
				t.Errorf("Label = %v, want %v", got.Label, tt.WantLabel)
			}
			if got.CatalogNumber != tt.WantCatalog {
				t.Errorf("CatalogNumber = %v, want %v", got.CatalogNumber, tt.WantCatalog)
			}
		})
	}
}
