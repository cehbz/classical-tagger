package scraping

import (
	"testing"
)

func TestLocalExtractor_ParseDirectoryName(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name      string
		dirPath   string
		wantTitle string
		wantYear  int
	}{
		{
			name:      "standard format with brackets",
			dirPath:   "/music/Beethoven - Symphony No. 5 [1963] [FLAC]",
			wantTitle: "Beethoven - Symphony No. 5",
			wantYear:  1963,
		},
		{
			name:      "parentheses for year",
			dirPath:   "/music/Bach - Goldberg Variations (1741)",
			wantTitle: "Bach - Goldberg Variations",
			wantYear:  1741,
		},
		{
			name:      "no year",
			dirPath:   "/music/Mozart - Piano Concertos [FLAC]",
			wantTitle: "Mozart - Piano Concertos",
			wantYear:  0,
		},
		{
			name:      "no format indicator",
			dirPath:   "/music/Vivaldi - Four Seasons [1989]",
			wantTitle: "Vivaldi - Four Seasons",
			wantYear:  1989,
		},
		{
			name:      "complex format with bit depth",
			dirPath:   "/music/J.S. Bach - Brandenburg Concertos [1982] [FLAC] [24-96]",
			wantTitle: "J.S. Bach - Brandenburg Concertos",
			wantYear:  1982,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotYear := extractor.parseDirectoryName(tt.dirPath)
			
			if gotTitle != tt.wantTitle {
				t.Errorf("parseDirectoryName() title = %v, want %v", gotTitle, tt.wantTitle)
			}
			if gotYear != tt.wantYear {
				t.Errorf("parseDirectoryName() year = %v, want %v", gotYear, tt.wantYear)
			}
		})
	}
}

func TestLocalExtractor_ExtractTrackNumberFromFilename(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name     string
		filename string
		want     int
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
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.extractTrackNumberFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("extractTrackNumberFromFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ExtractDiscFromPath(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name string
		path string
		want int
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
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.extractDiscFromPath(tt.path)
			if got != tt.want {
				t.Errorf("extractDiscFromPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ExtractTitleFromFilename(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"with track number", "/music/01 Symphony No. 5.flac", "Symphony No. 5"},
		{"with dash", "/music/01-Symphony No. 5.flac", "Symphony No. 5"},
		{"with dot", "/music/01.Symphony No. 5.flac", "Symphony No. 5"},
		{"no track number", "/music/Symphony No. 5.flac", "Symphony No. 5"},
		{"padded number", "/music/001 Concerto.flac", "Concerto"},
		{"with underscore", "/music/12_Piano Sonata.flac", "Piano Sonata"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.extractTitleFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("extractTitleFromFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_InferRoleFromName(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name       string
		artistName string
		want       string
	}{
		{"conductor explicit", "Herbert von Karajan, conductor", "conductor"},
		{"orchestra", "Berlin Philharmonic Orchestra", "ensemble"},
		{"philharmonic", "Vienna Philharmonic", "ensemble"},
		{"symphony", "London Symphony Orchestra", "ensemble"},
		{"choir", "RIAS Kammerchor", "ensemble"},
		{"chorus", "Westminster Choir", "ensemble"},
		{"quartet", "Emerson String Quartet", "ensemble"},
		{"trio", "Beaux Arts Trio", "ensemble"},
		{"chamber", "English Chamber Orchestra", "ensemble"},
		{"consort", "Gabrieli Consort", "ensemble"},
		{"soloist", "Maurizio Pollini", "soloist"},
		{"soloist name", "Anne-Sophie Mutter", "soloist"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.inferRoleFromName(tt.artistName)
			if got != tt.want {
				t.Errorf("inferRoleFromName(%q) = %v, want %v", tt.artistName, got, tt.want)
			}
		})
	}
}

func TestLocalExtractor_ParseArtistField(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name      string
		field     string
		wantCount int
		wantFirst string
		wantRole  string
	}{
		{
			name:      "semicolon separated",
			field:     "Pollini; Berlin Phil; Karajan",
			wantCount: 3,
			wantFirst: "Pollini",
			wantRole:  "soloist",
		},
		{
			name:      "comma separated",
			field:     "Pollini, Berlin Philharmonic, Karajan",
			wantCount: 3,
			wantFirst: "Pollini",
			wantRole:  "soloist",
		},
		{
			name:      "single artist",
			field:     "Maurizio Pollini",
			wantCount: 1,
			wantFirst: "Maurizio Pollini",
			wantRole:  "soloist",
		},
		{
			name:      "with ensemble",
			field:     "RIAS Kammerchor; Hans-Christoph Rademann",
			wantCount: 2,
			wantFirst: "RIAS Kammerchor",
			wantRole:  "ensemble",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.parseArtistField(tt.field)
			
			if len(got) != tt.wantCount {
				t.Errorf("parseArtistField(%q) returned %d artists, want %d", tt.field, len(got), tt.wantCount)
			}
			
			if len(got) > 0 {
				if got[0].Name != tt.wantFirst {
					t.Errorf("First artist name = %v, want %v", got[0].Name, tt.wantFirst)
				}
				if got[0].Role != tt.wantRole {
					t.Errorf("First artist role = %v, want %v", got[0].Role, tt.wantRole)
				}
			}
		})
	}
}

func TestLocalExtractor_ExtractEditionFromComment(t *testing.T) {
	extractor := NewLocalExtractor()
	
	tests := []struct {
		name        string
		comment     string
		wantLabel   string
		wantCatalog string
		wantNil     bool
	}{
		{
			name:        "label and catalog",
			comment:     "Label: Deutsche Grammophon\nCatalog: 479 1234",
			wantLabel:   "Deutsche Grammophon",
			wantCatalog: "479 1234",
			wantNil:     false,
		},
		{
			name:        "label only",
			comment:     "Label: Harmonia Mundi",
			wantLabel:   "Harmonia Mundi",
			wantCatalog: "",
			wantNil:     false,
		},
		{
			name:        "catalog only",
			comment:     "Catalog Number: HMC902170",
			wantLabel:   "",
			wantCatalog: "HMC902170",
			wantNil:     false,
		},
		{
			name:        "case insensitive",
			comment:     "LABEL: Test Label\nCATALOG: ABC123",
			wantLabel:   "Test Label",
			wantCatalog: "ABC123",
			wantNil:     false,
		},
		{
			name:    "no edition data",
			comment: "Just some random comment",
			wantNil: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.extractEditionFromComment(tt.comment)
			
			if tt.wantNil {
				if got != nil {
					t.Error("extractEditionFromComment() should return nil, got non-nil")
				}
				return
			}
			
			if got == nil {
				t.Fatal("extractEditionFromComment() returned nil, want non-nil")
			}
			
			if got.Label != tt.wantLabel {
				t.Errorf("Label = %v, want %v", got.Label, tt.wantLabel)
			}
			if got.CatalogNumber != tt.wantCatalog {
				t.Errorf("CatalogNumber = %v, want %v", got.CatalogNumber, tt.wantCatalog)
			}
		})
	}
}

// TestLocalExtractor_ImmutabilityPattern verifies that ExtractionResult is used immutably
func TestLocalExtractor_ImmutabilityPattern(t *testing.T) {	
	// Test that we can chain warnings/errors immutably
	data := &AlbumData{
		Title:        "Test Album",
		OriginalYear: 2020,
		Tracks:       []TrackData{},
	}
	
	result1 := NewExtractionResult(data)
	result2 := result1.WithWarning("test warning")
	result3 := result2.WithError(NewExtractionError("test", "test error", false))
	
	// Original should be unchanged
	if len(result1.Warnings()) != 0 {
		t.Error("result1 was mutated - should have 0 warnings")
	}
	if len(result1.Errors()) != 0 {
		t.Error("result1 was mutated - should have 0 errors")
	}
	
	// Second should have warning only
	if len(result2.Warnings()) != 1 {
		t.Error("result2 should have 1 warning")
	}
	if len(result2.Errors()) != 0 {
		t.Error("result2 should have 0 errors")
	}
	
	// Third should have both
	if len(result3.Warnings()) != 1 {
		t.Error("result3 should have 1 warning")
	}
	if len(result3.Errors()) != 1 {
		t.Error("result3 should have 1 error")
	}
}