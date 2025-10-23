package scraping

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// HarmoniaMundiParser parses Harmonia Mundi HTML pages.
type HarmoniaMundiParser struct {
	// Parser is stateless and immutable
}

// NewHarmoniaMundiParser creates a new parser instance.
func NewHarmoniaMundiParser() *HarmoniaMundiParser {
	return &HarmoniaMundiParser{}
}

// Parse parses a complete Harmonia Mundi HTML page and returns extraction result.
func (p *HarmoniaMundiParser) Parse(html string) (*ExtractionResult, error) {
	data := &AlbumData{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]TrackData, 0),
	}

	result := NewExtractionResult(data)
	parsingNotes := make(map[string]interface{})

	// Parse title
	if title, err := p.ParseTitle(html); err == nil && title != "" {
		data.Title = title
	} else {
		result = result.WithError(NewExtractionError("title", "not found in HTML", true))
	}

	// Parse year
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
	} else {
		result = result.WithError(NewExtractionError("year", "not found in HTML", true))
	}

	// Parse catalog number and label
	if catalog, err := p.ParseCatalogNumber(html); err == nil && catalog != "" {
		edition := &EditionData{
			Label:         "harmonia mundi", // Always harmonia mundi for this site
			CatalogNumber: catalog,
			EditionYear:   data.OriginalYear, // Use original year as edition year
		}
		data.Edition = edition
	} else {
		// Catalog is optional
		result = result.WithError(NewExtractionError("catalog_number", "not found in HTML", false))
	}

	// Parse artists (for notes)
	if artistText, err := p.ParseArtists(html); err == nil && artistText != "" {
		inferences := ParseArtistList(artistText)

		artistNotes := make([]map[string]interface{}, len(inferences))
		for i, inf := range inferences {
			artistNotes[i] = FormatInferenceForJSON(inf)

			// Warn on low confidence
			if IsLowConfidence(inf) {
				result = result.WithWarning(
					fmt.Sprintf("Low confidence artist inference: %s as %s (%s)",
						inf.ParsedName(), inf.InferredRole(), inf.Confidence()))
			}
		}
		parsingNotes["artists"] = artistNotes
	}

	// Parse tracks
	if tracks, err := p.ParseTracks(html); err == nil && len(tracks) > 0 {
		data.Tracks = tracks
	} else {
		result = result.WithError(NewExtractionError("tracks", "no tracks found in HTML", true))
	}

	// Add disc detection notes
	if len(data.Tracks) > 0 {
		trackLines := make([]string, len(data.Tracks))
		for i, track := range data.Tracks {
			trackLines[i] = fmt.Sprintf("%d. %s", track.Track, track.Title)
		}

		structure := DetectDiscStructure(trackLines)
		parsingNotes["disc_detection"] = map[string]interface{}{
			"disc_count":       structure.DiscCount(),
			"is_multi_disc":    structure.IsMultiDisc(),
			"detection_method": "track parsing",
		}
	}

	// Add parsing notes to result
	if len(parsingNotes) > 0 {
		result = result.WithParsingNotes(parsingNotes)
	}

	return result, nil
}

// ParseTitle extracts the album title from the HTML title tag.
func (p *HarmoniaMundiParser) ParseTitle(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get title from <title> tag
	title := doc.Find("title").First().Text()
	if title == "" {
		return "", fmt.Errorf("no title tag found")
	}

	// Remove " | harmonia mundi" suffix
	title = strings.TrimSpace(title)
	if idx := strings.LastIndex(title, " | harmonia mundi"); idx > 0 {
		title = title[:idx]
	}

	// Clean up HTML entities
	title = cleanHTMLEntities(title)

	return strings.TrimSpace(title), nil
}

// ParseYear extracts the year from the datePublished field in JSON-LD.
func (p *HarmoniaMundiParser) ParseYear(html string) (int, error) {
	// Look for datePublished in JSON-LD structured data
	re := regexp.MustCompile(`"datePublished":\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return 0, fmt.Errorf("datePublished not found")
	}

	dateStr := matches[1]

	// Extract year from date string (e.g., "October 2013" or "2013-10-01")
	yearRe := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	yearMatches := yearRe.FindString(dateStr)

	if yearMatches == "" {
		return 0, fmt.Errorf("no year found in date: %s", dateStr)
	}

	year, err := strconv.Atoi(yearMatches)
	if err != nil {
		return 0, fmt.Errorf("failed to parse year: %w", err)
	}

	return year, nil
}

// ParseCatalogNumber extracts the catalog number.
func (p *HarmoniaMundiParser) ParseCatalogNumber(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Look for catalog number in <div class="feature ref">
	catalog := doc.Find(".feature.ref").First().Text()
	catalog = strings.TrimSpace(catalog)

	if catalog == "" {
		return "", fmt.Errorf("catalog number not found")
	}

	return catalog, nil
}

// ParseArtists extracts the main artist/ensemble from JSON-LD.
func (p *HarmoniaMundiParser) ParseArtists(html string) (string, error) {
	// Look for byArtist in JSON-LD
	re := regexp.MustCompile(`"byArtist":\s*\{[^}]*"name":\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return "", fmt.Errorf("byArtist not found")
	}

	return matches[1], nil
}

// ParseTracks extracts track listings from the HTML.
func (p *HarmoniaMundiParser) ParseTracks(html string) ([]TrackData, error) {
	// The track listing is in plain text with <br> tags
	// Pattern: COMPOSER [dates]<br>· <b>Title</b> (timing)<br>

	tracks := make([]TrackData, 0)
	currentComposer := ""
	trackNum := 0
	disc := 1

	// Split by <br> tags
	lines := strings.Split(html, "<br>")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a composer line (all caps, with or without dates)
		if isComposerLine(line) {
			composer := extractComposerName(line)
			if composer != "" {
				currentComposer = composer
			}
			continue
		}

		// Check if this is a track line (starts with ·)
		if strings.HasPrefix(line, "·") || strings.HasPrefix(line, "Â·") {
			if currentComposer == "" {
				currentComposer = MissingComposer
			}

			trackNum++

			// Extract title from <b> tags
			title := extractTextFromBold(line)
			if title == "" {
				title = MissingTrackTitle
			}

			track := TrackData{
				Disc:     disc,
				Track:    trackNum,
				Title:    cleanHTMLEntities(title),
				Composer: currentComposer,
				Artists:  make([]ArtistData, 0),
			}

			tracks = append(tracks, track)
		}
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}

// isComposerLine checks if a line is a composer name line.
func isComposerLine(line string) bool {
	// Remove HTML tags for checking
	line = stripHTMLTags(line)
	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return false
	}

	// Composer lines have dates in brackets, e.g., [1809-1847]
	if strings.Contains(line, "[") && strings.Contains(line, "]") {
		return true
	}

	// OR: All uppercase line that doesn't start with · (track indicator)
	// Check if it's mostly uppercase (composer names are ALL CAPS)
	if strings.HasPrefix(line, "·") || strings.HasPrefix(line, "Â·") {
		return false // This is a track line
	}

	upperCount := 0
	letterCount := 0
	for _, r := range line {
		if unicode.IsLetter(r) {
			letterCount++
			if unicode.IsUpper(r) {
				upperCount++
			}
		}
	}

	// If more than 80% uppercase letters, it's likely a composer
	// (handles cases like "UWE GRONOSTAY" without dates)
	if letterCount > 0 && float64(upperCount)/float64(letterCount) > 0.8 {
		return true
	}

	return false
}

// extractComposerName extracts the composer name from a composer line.
func extractComposerName(line string) string {
	// Remove HTML tags
	line = stripHTMLTags(line)
	line = strings.TrimSpace(line)

	// Find the part before the dates [...] if present
	if idx := strings.Index(line, "["); idx > 0 {
		line = line[:idx]
		line = strings.TrimSpace(line)
	}

	// Convert from ALL CAPS to Title Case
	return toTitleCase(line)
}

// extractTextFromBold extracts text from <b> tags.
func extractTextFromBold(line string) string {
	re := regexp.MustCompile(`<b>([^<]+)</b>`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}
