package scraping

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cehbz/classical-tagger/internal/domain"
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
	data := &domain.Album{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]*domain.Track, 0),
	}

	result := &ExtractionResult{
		Album: data,
		Source: "discogs",
	}
	parsingNotes := make(map[string]interface{})
	parsingNotes["source"] = "discogs"

	// Parse title from JSON-LD
	if title, err := p.ParseTitle(html); err == nil && title != "" {
		data.Title = title
	} else {
		result.Warnings = append(result.Warnings, "title not found in JSON-LD")
	}

	// Parse year from JSON-LD
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
	} else {
		result.Warnings = append(result.Warnings, "year not found in JSON-LD")
	}

	// Parse catalog number and label from JSON-LD
	catalog, catalogErr := p.ParseCatalogNumber(html)
	label, labelErr := p.ParseLabel(html)

	if catalogErr == nil || labelErr == nil {
		edition := &domain.Edition{
			Label:         label,
			CatalogNumber: catalog,
			Year:          data.OriginalYear,
		}
		data.Edition = edition
	} else {
		result.Warnings = append(result.Warnings, "catalog number or label not found in JSON-LD")
	}

	// Parse tracks from table structure
	if tracks, err := p.ParseTracks(html); err == nil && len(tracks) > 0 {
		data.Tracks = tracks
		parsingNotes["tracks_source"] = "tracklist_table"
	} else {
		result.Warnings = append(result.Warnings, "no tracks found in HTML")
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
		for key, value := range parsingNotes {
			result.Notes = append(result.Notes, fmt.Sprintf("%s: %v", key, value))
		}
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
func (p *DiscogsParser) ParseTracks(html string) ([]*domain.Track, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	tracks := make([]*domain.Track, 0)
	var currentHeadingComposer string
	var currentHeadingTitle string

	// Parse tracklist table
	doc.Find("table.tracklist_ZdQ0I tbody tr").Each(func(i int, row *goquery.Selection) {
		// Check if this is a heading row (for multi-movement works)
		if row.HasClass("heading_mkZNt") {
			// Extract parent work title from heading
			titleCell := row.Find(".trackTitle_loyWF")
			// Clone to avoid modifying original
			titleClone := titleCell.Clone()
			// Remove credits div to get just the title
			titleClone.Find(".credits_vzBtg").Remove()
			currentHeadingTitle = strings.TrimSpace(titleClone.Text())
			currentHeadingTitle = cleanHTMLEntities(currentHeadingTitle)

			// Extract composer from heading if available
			creditsDiv := row.Find(".credits_vzBtg")
			if creditsDiv.Length() > 0 {
				composerLink := creditsDiv.Find("a[href*='artist']").First()
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
			titleCellClone := titleCell.Clone()
			titleCellClone.Find(".credits_vzBtg").Remove()
			title = strings.TrimSpace(titleCellClone.Text())
		}
		title = cleanHTMLEntities(title)

		// Prepend parent work title for subtracks
		if isSubtrack && currentHeadingTitle != "" {
			title = currentHeadingTitle + ": " + title
		}

		// Extract composer
		composer := ""
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

		// Reset heading context when we encounter a non-subtrack
		if !isSubtrack {
			currentHeadingTitle = ""
			currentHeadingComposer = ""
		}

		if title != "" {
			track := &domain.Track{
				Disc:     1,
				Track:    trackNum,
				Title:    title,
				Artists:  []domain.Artist{domain.Artist{Name: composer, Role: domain.RoleComposer}},
			}
			tracks = append(tracks, track)
		}
	})

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}
