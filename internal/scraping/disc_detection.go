package scraping

import (
	"regexp"
	"strconv"
	"strings"
)

// DiscStructure represents the disc structure of an album.
// It is immutable after creation.
type DiscStructure struct {
	discCount  int
	lineToDisc map[int]int  // Maps line index to disc number
	trackLines map[int]bool // Maps line index to whether it's a track line
}

// DetectDiscStructure analyzes lines of text and determines disc structure.
func DetectDiscStructure(lines []string) *DiscStructure {
	if len(lines) == 0 {
		return &DiscStructure{
			discCount:  1,
			lineToDisc: make(map[int]int),
			trackLines: make(map[int]bool),
		}
	}
	
	lineToDisc := make(map[int]int)
	trackLines := make(map[int]bool)
	currentDisc := 1
	maxDisc := 1
	lastTrackNum := 0
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for explicit disc headers
		if IsDiscHeader(line) {
			if disc, ok := ExtractDiscNumber(line); ok {
				currentDisc = disc
				if disc > maxDisc {
					maxDisc = disc
				}
				lineToDisc[i] = currentDisc
				trackLines[i] = false // Header line, not a track
				lastTrackNum = 0      // Reset track counter
				continue
			}
		}
		
		// Check for track lines
		if trackNum := extractTrackNumber(line); trackNum > 0 {
			// Detect track number reset (indicates new disc)
			if trackNum == 1 && lastTrackNum > 1 {
				currentDisc++
				if currentDisc > maxDisc {
					maxDisc = currentDisc
				}
			}
			lastTrackNum = trackNum
			lineToDisc[i] = currentDisc
			trackLines[i] = true
		} else if line != "" {
			// Non-empty, non-track line
			lineToDisc[i] = currentDisc
			trackLines[i] = false
		}
	}
	
	return &DiscStructure{
		discCount:  maxDisc,
		lineToDisc: lineToDisc,
		trackLines: trackLines,
	}
}

// DiscCount returns the number of discs detected.
func (d *DiscStructure) DiscCount() int {
	return d.discCount
}

// IsMultiDisc returns true if more than one disc was detected.
func (d *DiscStructure) IsMultiDisc() bool {
	return d.discCount > 1
}

// GetDiscNumber returns the disc number for a given line index.
// Returns 1 if line index not found.
func (d *DiscStructure) GetDiscNumber(lineIndex int) int {
	if disc, ok := d.lineToDisc[lineIndex]; ok {
		return disc
	}
	return 1
}

// IsTrackLine returns true if the given line index is a track line.
func (d *DiscStructure) IsTrackLine(lineIndex int) bool {
	return d.trackLines[lineIndex]
}

// IsDiscHeader checks if a line is a disc header (e.g., "CD 1", "Disc 2").
func IsDiscHeader(line string) bool {
	line = strings.TrimSpace(line)
	_, ok := ExtractDiscNumber(line)
	return ok
}

// ExtractDiscNumber extracts a disc number from a header line.
// Returns (discNumber, true) if found, (0, false) otherwise.
func ExtractDiscNumber(line string) (int, bool) {
	line = strings.TrimSpace(line)
	lower := strings.ToLower(line)
	
	// Pattern: "CD 1", "CD1", "Disc 1", "1:", etc.
	patterns := []struct {
		prefix string
		re     *regexp.Regexp
	}{
		{"cd", regexp.MustCompile(`^cd\s*(\d+)$`)},
		{"disc", regexp.MustCompile(`^disc\s*(\d+)$`)},
		{"", regexp.MustCompile(`^(\d+):$`)},
	}
	
	for _, p := range patterns {
		if p.prefix == "" || strings.HasPrefix(lower, p.prefix) {
			if matches := p.re.FindStringSubmatch(lower); len(matches) > 1 {
				if num, err := strconv.Atoi(matches[1]); err == nil && num > 0 {
					return num, true
				}
			}
		}
	}
	
	return 0, false
}

// extractTrackNumber extracts a track number from a track line.
// Returns 0 if no track number found.
func extractTrackNumber(line string) int {
	line = strings.TrimSpace(line)
	
	// Pattern: "1. Title", "Track 1: Title", "01 Title"
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^(\d+)\.\s`),           // "1. Title"
		regexp.MustCompile(`^Track\s+(\d+)[\s:]`),  // "Track 1: Title"
		regexp.MustCompile(`^(\d+)\s+[A-Z]`),       // "01 Title" (letter after space)
		regexp.MustCompile(`^·\s*(\d+)\s`),         // "· 1 Title" (bullet format)
	}
	
	for _, re := range patterns {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return num
			}
		}
	}
	
	return 0
}

// DetectDiscStructureFromHTML is a convenience function for HTML parsing.
// It extracts text content and delegates to DetectDiscStructure.
func DetectDiscStructureFromHTML(htmlLines []string) *DiscStructure {
	// Remove HTML tags and clean up
	cleaned := make([]string, 0, len(htmlLines))
	for _, line := range htmlLines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return DetectDiscStructure(cleaned)
}