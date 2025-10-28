package filesystem

import "testing"

func TestIsDiscDirectory(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		// Happy path - standard formats
		{"CD with number", "CD1", true},
		{"CD with space", "CD 1", true},
		{"lowercase cd", "cd1", true},
		{"Disc with number", "Disc1", true},
		{"Disc with space", "Disc 2", true},
		{"lowercase disc", "disc1", true},
		{"Disk variant", "Disk1", true},
		{"DVD format", "DVD1", true},
		
		// Edge case - prefix only (no number)
		{"CD alone", "CD", true},
		{"Disc alone", "Disc", true},
		{"lowercase cd alone", "cd", true},
		
		// Happy path - double digits
		{"CD10", "CD10", true},
		{"Disc99", "Disc99", true},
		
		// Happy path - leading/trailing spaces
		{"leading space", "  CD1", true},
		{"trailing space", "CD1  ", true},
		{"both spaces", "  Disc 2  ", true},
		
		// Failure cases - not disc directories
		{"Artist name", "Artist", false},
		{"Album name", "Album", false},
		{"Year", "1963", false},
		{"Beethoven", "Beethoven", false},
		{"empty string", "", false},
		{"only spaces", "   ", false},
		
		// Failure cases - prefix with non-digits
		{"CD with text", "CDextra", false},
		{"Disc with text", "Discotheque", false},
		{"CD with hyphen", "CD-ROM", false},
		{"mixed", "CD1a", false},
		{"CD with underscore", "CD_1", false},
		
		// Edge cases - case variations
		{"uppercase", "CD1", true},
		{"lowercase", "cd1", true},
		{"mixed case", "Cd1", true},
		{"DISC uppercase", "DISC2", true},
		
		// Edge cases - whitespace variations
		{"CD space 1", "CD 1", true},
		{"CD multiple spaces", "CD  1", true},
		{"CD tab", "CD\t1", true},
		
		// Edge cases - three digit numbers
		{"CD100", "CD100", true},
		{"Disc999", "Disc999", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDiscDirectory(tt.dir)
			if got != tt.want {
				t.Errorf("IsDiscDirectory(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// TestIsDiscDirectory_NestedPaths tests that the function works on directory names,
// not full paths (caller's responsibility to extract basename)
func TestIsDiscDirectory_NestedPaths(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		basename string
		want     bool
	}{
		{
			name:     "nested CD1",
			path:     "/music/Album/CD1/file.flac",
			basename: "CD1",
			want:     true,
		},
		{
			name:     "nested Disc 2",
			path:     "/music/Artist/Album/Disc 2/01.flac",
			basename: "Disc 2",
			want:     true,
		},
		{
			name:     "nested non-disc",
			path:     "/music/Album/Extras/file.flac",
			basename: "Extras",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The function should be called with basename only
			got := IsDiscDirectory(tt.basename)
			if got != tt.want {
				t.Errorf("IsDiscDirectory(%q) = %v, want %v (from path %q)",
					tt.basename, got, tt.want, tt.path)
			}
		})
	}
}

// TestIsDiscDirectory_RealWorldExamples tests actual directory names from classical torrents
func TestIsDiscDirectory_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		// Real examples from torrents
		{"standard CD1", "CD1", true},
		{"standard CD2", "CD2", true},
		{"Disc 1", "Disc 1", true},
		{"Disc 2", "Disc 2", true},
		{"DVD1", "DVD1", true},
		
		// Non-disc directories from real torrents
		{"Booklet", "Booklet", false},
		{"Covers", "Covers", false},
		{"Scans", "Scans", false},
		{"Artwork", "Artwork", false},
		{"Bonus", "Bonus", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDiscDirectory(tt.dir)
			if got != tt.want {
				t.Errorf("IsDiscDirectory(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// BenchmarkIsDiscDirectory measures performance
func BenchmarkIsDiscDirectory(b *testing.B) {
	testCases := []string{
		"CD1",
		"Disc 2",
		"Artist",
		"CDextra",
		"cd10",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			IsDiscDirectory(tc)
		}
	}
}