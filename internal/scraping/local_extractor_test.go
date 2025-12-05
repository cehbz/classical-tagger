package scraping

import (
	"testing"
)

func TestParseDirectoryName(t *testing.T) {
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
			gotFolder, gotTitle, gotYear := parseDirectoryName(tt.DirPath)

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

func TestExtractTrackNumberFromFilename(t *testing.T) {
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
			got := extractTrackNumberFromFilename(tt.Filename)
			if got != tt.Want {
				t.Errorf("extractTrackNumberFromFilename(%q) = %v, want %v", tt.Filename, got, tt.Want)
			}
		})
	}
}

func TestExtractDiscFromPath(t *testing.T) {
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
			got := extractDiscFromPath(tt.Path)
			if got != tt.Want {
				t.Errorf("extractDiscFromPath(%q) = %v, want %v", tt.Path, got, tt.Want)
			}
		})
	}
}

func TestExtractTitleFromFilename(t *testing.T) {
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
			got := extractTitleFromFilename(tt.Filename)
			if got != tt.Want {
				t.Errorf("extractTitleFromFilename(%q) = %v, want %v", tt.Filename, got, tt.Want)
			}
		})
	}
}

func TestExtractEditionFromTags(t *testing.T) {
	tests := []struct {
		Name        string
		Tags        map[string]string
		WantLabel   string
		WantCatalog string
		WantYear    int
		WantNil     bool
	}{
		{
			Name: "label, catalog, and date",
			Tags: map[string]string{
				"LABEL":         "Deutsche Grammophon",
				"CATALOGNUMBER": "479 1234",
				"DATE":          "1992",
			},
			WantLabel:   "Deutsche Grammophon",
			WantCatalog: "479 1234",
			WantYear:    1992,
			WantNil:     false,
		},
		{
			Name: "label only",
			Tags: map[string]string{
				"LABEL": "Harmonia Mundi",
			},
			WantLabel:   "Harmonia Mundi",
			WantCatalog: "",
			WantYear:    0,
			WantNil:     false,
		},
		{
			Name: "catalog only",
			Tags: map[string]string{
				"CATALOGNUMBER": "HMC902170",
			},
			WantLabel:   "",
			WantCatalog: "HMC902170",
			WantYear:    0,
			WantNil:     false,
		},
		{
			Name: "date only",
			Tags: map[string]string{
				"DATE": "2013",
			},
			WantLabel:   "",
			WantCatalog: "",
			WantYear:    2013,
			WantNil:     false,
		},
		{
			Name: "label and date",
			Tags: map[string]string{
				"LABEL": "Sony Classical",
				"DATE":  "1992",
			},
			WantLabel:   "Sony Classical",
			WantCatalog: "",
			WantYear:    1992,
			WantNil:     false,
		},
		{
			Name: "no edition tags",
			Tags: map[string]string{
				"TITLE":  "Some Title",
				"ARTIST": "Some Artist",
			},
			WantNil: true,
		},
		{
			Name:    "empty tags",
			Tags:    map[string]string{},
			WantNil: true,
		},
		{
			Name: "invalid date",
			Tags: map[string]string{
				"LABEL": "Test Label",
				"DATE":  "invalid",
			},
			WantLabel:   "Test Label",
			WantCatalog: "",
			WantYear:    0,
			WantNil:     false,
		},
		{
			Name: "date with whitespace",
			Tags: map[string]string{
				"DATE": "  2013  ",
			},
			WantLabel:   "",
			WantCatalog: "",
			WantYear:    2013,
			WantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := extractEditionFromTags(tt.Tags)

			if tt.WantNil {
				if got != nil {
					t.Error("extractEditionFromTags() should return nil, got non-nil")
				}
				return
			}

			if got == nil {
				t.Fatal("extractEditionFromTags() returned nil, want non-nil")
			}

			if got.Label != tt.WantLabel {
				t.Errorf("Label = %v, want %v", got.Label, tt.WantLabel)
			}
			if got.CatalogNumber != tt.WantCatalog {
				t.Errorf("CatalogNumber = %v, want %v", got.CatalogNumber, tt.WantCatalog)
			}
			if got.Year != tt.WantYear {
				t.Errorf("Year = %v, want %v", got.Year, tt.WantYear)
			}
		})
	}
}

func TestExtractEditionFromComment(t *testing.T) {
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
			got := extractEditionFromComment(tt.Comment)

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
