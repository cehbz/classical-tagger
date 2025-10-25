package scraping

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DiscogsParser parses Discogs HTML pages.
type DiscogsParser struct {
	// Parser is stateless and immutable
}

// NewDiscogsParser creates a new parser instance.
func NewDiscogsParser() *DiscogsParser {
	return &DiscogsParser{}
}

// Parse parses a complete Discogs HTML page and returns extraction result.
func (p *DiscogsParser) Parse(html string) (*ExtractionResult, error) {
	data := &AlbumData{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]TrackData, 0),
	}

	result := NewExtractionResult(data)
	parsingNotes := make(map[string]interface{})
	parsingNotes["source"] = "discogs"

	// Parse title from JSON-LD
	if title, err := p.ParseTitle(html); err == nil && title != "" {
		data.Title = title
	} else {
		result = result.WithError(NewExtractionError("title", "not found in JSON-LD", true))
	}

	// Parse year from JSON-LD
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
	} else {
		result = result.WithError(NewExtractionError("year", "not found in JSON-LD", true))
	}

	// Parse catalog number and label from JSON-LD
	catalog, catalogErr := p.ParseCatalogNumber(html)
	label, labelErr := p.ParseLabel(html)

	if catalogErr == nil || labelErr == nil {
		edition := &EditionData{
			Label:         label,
			CatalogNumber: catalog,
			EditionYear:   data.OriginalYear,
		}
		data.Edition = edition
	} else {
		result = result.WithError(NewExtractionError("catalog_number", "not found in JSON-LD", false))
	}

	// Parse tracks from table structure
	if tracks, err := p.ParseTracks(html); err == nil && len(tracks) > 0 {
		data.Tracks = tracks
		parsingNotes["tracks_source"] = "tracklist_table"
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

// ParseTitle extracts the album title from JSON-LD structured data.
func (p *DiscogsParser) ParseTitle(html string) (string, error) {
	// Look for JSON-LD script tag with id="release_schema"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	jsonLD := doc.Find("script#release_schema[type='application/ld+json']").First().Text()
	if jsonLD == "" {
		return "", fmt.Errorf("no JSON-LD release_schema found")
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLD), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	// Extract name field
	if name, ok := data["name"].(string); ok && name != "" {
		return cleanHTMLEntities(name), nil
	}

	return "", fmt.Errorf("no name field in JSON-LD")
}

// ParseYear extracts the year from JSON-LD structured data.
func (p *DiscogsParser) ParseYear(html string) (int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %w", err)
	}

	jsonLD := doc.Find("script#release_schema[type='application/ld+json']").First().Text()
	if jsonLD == "" {
		return 0, fmt.Errorf("no JSON-LD release_schema found")
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLD), &data); err != nil {
		return 0, fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	// Extract datePublished - can be number or string
	if datePublished, ok := data["datePublished"]; ok {
		switch v := datePublished.(type) {
		case float64:
			return int(v), nil
		case string:
			year := 0
			if _, err := fmt.Sscanf(v, "%d", &year); err == nil && year > 0 {
				return year, nil
			}
		}
	}

	return 0, fmt.Errorf("no datePublished field in JSON-LD")
}

// ParseCatalogNumber extracts the catalog number from JSON-LD.
func (p *DiscogsParser) ParseCatalogNumber(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	jsonLD := doc.Find("script#release_schema[type='application/ld+json']").First().Text()
	if jsonLD == "" {
		return "", fmt.Errorf("no JSON-LD release_schema found")
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLD), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	// Extract catalogNumber field
	if catalogNumber, ok := data["catalogNumber"].(string); ok && catalogNumber != "" {
		return strings.TrimSpace(catalogNumber), nil
	}

	return "", fmt.Errorf("no catalogNumber field in JSON-LD")
}

// ParseLabel extracts the label from JSON-LD.
func (p *DiscogsParser) ParseLabel(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	jsonLD := doc.Find("script#release_schema[type='application/ld+json']").First().Text()
	if jsonLD == "" {
		return "", fmt.Errorf("no JSON-LD release_schema found")
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLD), &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	// Extract recordLabel - can be array or object
	if recordLabel, ok := data["recordLabel"]; ok {
		switch labels := recordLabel.(type) {
		case []interface{}:
			if len(labels) > 0 {
				if labelObj, ok := labels[0].(map[string]interface{}); ok {
					if name, ok := labelObj["name"].(string); ok && name != "" {
						return strings.TrimSpace(name), nil
					}
				}
			}
		case map[string]interface{}:
			if name, ok := labels["name"].(string); ok && name != "" {
				return strings.TrimSpace(name), nil
			}
		}
	}

	return "", fmt.Errorf("no recordLabel field in JSON-LD")
}

// ParseTracks extracts track listings from the Discogs tracklist table.
func (p *DiscogsParser) ParseTracks(html string) ([]TrackData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	tracks := make([]TrackData, 0)
	var currentHeadingComposer string

	// Parse tracklist table
	doc.Find("table.tracklist_ZdQ0I tbody tr").Each(func(i int, row *goquery.Selection) {
		// Check if this is a heading row (for multi-movement works)
		if row.HasClass("heading_mkZNt") {
			// Extract composer from heading if available
			creditsDiv := row.Find(".credits_vzBtg")
			if creditsDiv.Length() > 0 {
				composerLink := creditsDiv.Find("a[href*='artist']")
				if composerLink.Length() > 0 {
					currentHeadingComposer = strings.TrimSpace(composerLink.Text())
					currentHeadingComposer = cleanHTMLEntities(currentHeadingComposer)
				}
			}
			return
		}

		// Check if this is a subtrack row
		isSubtrack := row.HasClass("subtrack_o3GgI")

		// Get track position attribute
		position, posExists := row.Attr("data-track-position")
		if !posExists && !isSubtrack {
			return // Skip rows without position
		}

		var trackNum int
		if isSubtrack {
			// Extract track number from subtrack position cell
			posText := strings.TrimSpace(row.Find(".subtrackPos_HC1me").Text())
			// Remove any icons/symbols
			posText = regexp.MustCompile(`[^0-9]`).ReplaceAllString(posText, "")
			fmt.Sscanf(posText, "%d", &trackNum)
		} else {
			fmt.Sscanf(position, "%d", &trackNum)
		}

		if trackNum == 0 {
			trackNum = len(tracks) + 1
		}

		// Extract title
		titleCell := row.Find(".trackTitle_loyWF")
		titleSpan := titleCell.Find("span").First()
		title := strings.TrimSpace(titleSpan.Text())
		if title == "" {
			// Fallback: get all text from title cell, excluding credits
			titleCell.Find(".credits_vzBtg").Remove()
			title = strings.TrimSpace(titleCell.Text())
		}
		title = cleanHTMLEntities(title)

		// FIXED CODE - extracts once from link only
		composer := ""
		// Find the credits div
		creditDiv := row.Find(".credits_vzBtg")
		if creditDiv.Length() > 0 {
			// Only extract from the composer link, not surrounding text
			composerLink := creditDiv.Find("a[href*='/artist/']").First()
			if composerLink.Length() > 0 {
				composer = strings.TrimSpace(composerLink.Text())
				composer = cleanHTMLEntities(composer)
			}
		}

		// Use heading composer for subtracks if no specific composer found
		if composer == "" && isSubtrack && currentHeadingComposer != "" {
			composer = currentHeadingComposer
		}

		if title != "" {
			track := TrackData{
				Disc:     1,
				Track:    trackNum,
				Title:    title,
				Composer: composer,
			}
			tracks = append(tracks, track)
		}
	})

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}