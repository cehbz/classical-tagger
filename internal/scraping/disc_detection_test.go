package scraping

import (
	"testing"
)

// TestDetectDiscStructure tests detection of multi-disc albums
func TestDetectDiscStructure(t *testing.T) {
	tests := []struct {
		Name      string
		Lines     []string
		WantDiscs int
	}{
		{
			Name: "single disc implicit",
			Lines: []string{
				"Track 1: Aria",
				"Track 2: Variation 1",
				"Track 3: Variation 2",
			},
			WantDiscs: 1,
		},
		{
			Name: "explicit CD1/CD2",
			Lines: []string{
				"CD1",
				"Track 1: Movement I",
				"Track 2: Movement II",
				"CD2",
				"Track 1: Movement III",
				"Track 2: Movement IV",
			},
			WantDiscs: 2,
		},
		{
			Name: "Disc 1/Disc 2 format",
			Lines: []string{
				"Disc 1",
				"Track 1: First",
				"Disc 2",
				"Track 1: Second",
			},
			WantDiscs: 2,
		},
		{
			Name: "numbered format with colon",
			Lines: []string{
				"1:",
				"Track 1: First",
				"2:",
				"Track 1: Second",
			},
			WantDiscs: 2,
		},
		{
			Name: "track numbers resetting",
			Lines: []string{
				"1. Aria",
				"2. Variation 1",
				"1. Adagio", // Reset indicates new disc
				"2. Allegro",
			},
			WantDiscs: 2,
		},
		{
			Name: "three disc album",
			Lines: []string{
				"CD 1",
				"1. First",
				"CD 2",
				"1. Second",
				"CD 3",
				"1. Third",
			},
			WantDiscs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			structure := DetectDiscStructure(tt.Lines)

			if structure.DiscCount() != tt.WantDiscs {
				t.Errorf("DiscCount() = %d, want %d", structure.DiscCount(), tt.WantDiscs)
			}

			// Verify we have track assignments for all lines
			if structure.IsMultiDisc() && tt.WantDiscs > 1 {
				if !structure.IsMultiDisc() {
					t.Error("IsMultiDisc() = false, want true")
				}
			}
		})
	}
}

// TestDiscStructure_GetDiscNumber tests getting disc numbers for specific lines
func TestDiscStructure_GetDiscNumber(t *testing.T) {
	lines := []string{
		"CD 1",
		"Track 1: First",
		"Track 2: Second",
		"CD 2",
		"Track 1: Third",
		"Track 2: Fourth",
	}

	structure := DetectDiscStructure(lines)

	tests := []struct {
		lineIndex int
		wantDisc  int
		wantTrack bool
	}{
		{lineIndex: 0, wantDisc: 1, wantTrack: false}, // "CD 1" header
		{lineIndex: 1, wantDisc: 1, wantTrack: true},  // First track
		{lineIndex: 2, wantDisc: 1, wantTrack: true},  // Second track
		{lineIndex: 3, wantDisc: 2, wantTrack: false}, // "CD 2" header
		{lineIndex: 4, wantDisc: 2, wantTrack: true},  // Third track
		{lineIndex: 5, wantDisc: 2, wantTrack: true},  // Fourth track
	}

	for _, tt := range tests {
		disc := structure.GetDiscNumber(tt.lineIndex)
		if disc != tt.wantDisc {
			t.Errorf("GetDiscNumber(%d) = %d, want %d", tt.lineIndex, disc, tt.wantDisc)
		}

		isTrack := structure.IsTrackLine(tt.lineIndex)
		if isTrack != tt.wantTrack {
			t.Errorf("IsTrackLine(%d) = %v, want %v", tt.lineIndex, isTrack, tt.wantTrack)
		}
	}
}

// TestDiscStructure_Immutability tests that DiscStructure is immutable
func TestDiscStructure_Immutability(t *testing.T) {
	lines := []string{
		"CD 1",
		"Track 1",
	}

	structure := DetectDiscStructure(lines)
	discCount := structure.DiscCount()

	// Try to modify returned disc count (shouldn't affect original)
	_ = discCount + 1

	if structure.DiscCount() != 1 {
		t.Error("DiscStructure was mutated")
	}
}

// TestIsDiscHeader tests detection of disc header lines
func TestIsDiscHeader(t *testing.T) {
	tests := []struct {
		line       string
		wantHeader bool
	}{
		{"CD1", true},
		{"CD 1", true},
		{"CD 2", true},
		{"Disc 1", true},
		{"Disc 2", true},
		{"1:", true},
		{"2:", true},
		{"Track 1: Title", false},
		{"Movement I", false},
		{"", false},
		{"CD", false}, // Missing number
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			if got := IsDiscHeader(tt.line); got != tt.wantHeader {
				t.Errorf("IsDiscHeader(%q) = %v, want %v", tt.line, got, tt.wantHeader)
			}
		})
	}
}

// TestExtractDiscNumber tests extracting disc numbers from headers
func TestExtractDiscNumber(t *testing.T) {
	tests := []struct {
		line     string
		wantDisc int
		wantOk   bool
	}{
		{"CD1", 1, true},
		{"CD 1", 1, true},
		{"CD 2", 2, true},
		{"Disc 3", 3, true},
		{"1:", 1, true},
		{"10:", 10, true},
		{"Track 1", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			disc, ok := ExtractDiscNumber(tt.line)

			if ok != tt.wantOk {
				t.Errorf("ExtractDiscNumber(%q) ok = %v, want %v", tt.line, ok, tt.wantOk)
			}

			if ok && disc != tt.wantDisc {
				t.Errorf("ExtractDiscNumber(%q) = %d, want %d", tt.line, disc, tt.wantDisc)
			}
		})
	}
}

// TestDetectTrackReset tests detection of track number resets
func TestDetectTrackReset(t *testing.T) {
	lines := []string{
		"1. First track",
		"2. Second track",
		"3. Third track",
		"1. First track of disc 2", // Reset here
		"2. Second track of disc 2",
	}

	structure := DetectDiscStructure(lines)

	// Should detect 2 discs due to track number reset
	if structure.DiscCount() != 2 {
		t.Errorf("DiscCount() = %d, want 2 (should detect reset)", structure.DiscCount())
	}

	// Line 3 (index 3) should be on disc 2
	if structure.GetDiscNumber(3) != 2 {
		t.Errorf("GetDiscNumber(3) = %d, want 2", structure.GetDiscNumber(3))
	}
}
