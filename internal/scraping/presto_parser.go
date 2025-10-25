package scraping

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// PrestoParser parses Presto Classical HTML pages.
type PrestoParser struct {
	// Parser is stateless and immutable
}

// NewPrestoParser creates a new parser instance.
func NewPrestoParser() *PrestoParser {
	return &PrestoParser{}
}

// Parse parses a complete Presto Classical HTML page and returns extraction result.
func (p *PrestoParser) Parse(html string) (*ExtractionResult, error) {
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

	// Parse year from meta tag
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
	} else {
		result = result.WithError(NewExtractionError("year", "not found in HTML", true))
	}

	// Parse catalog number and label
	if catalog, label, err := p.ParseCatalogAndLabel(html); err == nil {
		edition := &EditionData{
			Label:         label,
			CatalogNumber: catalog,
			EditionYear:   data.OriginalYear,
		}
		data.Edition = edition
	} else {
		result = result.WithError(NewExtractionError("catalog_number", "not found in HTML", false))
	}

	// Parse tracks using semantic structure
	if tracks, err := p.ParseTracks(html); err == nil && len(tracks) > 0 {
		data.Tracks = tracks
		parsingNotes["tracks_source"] = "semantic_structure"
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
func (p *PrestoParser) ParseTitle(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get title from <title> tag
	title := doc.Find("title").First().Text()
	if title == "" {
		return "", fmt.Errorf("no title tag found")
	}

	// Remove " | Presto Music" suffix first if present
	if idx := strings.Index(title, " | Presto Music"); idx > 0 {
		title = title[:idx]
	}

	// Remove everything after the FIRST " - " (which contains label/catalog info)
	if idx := strings.Index(title, " - "); idx > 0 {
		title = title[:idx]
	}

	// Clean up HTML entities
	title = cleanHTMLEntities(title)

	return strings.TrimSpace(title), nil
}

// ParseYear extracts the year from the release date.
func (p *PrestoParser) ParseYear(html string) (int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %w", err)
	}

	yearRe := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	
	// Look for release date in product metadata
	doc.Find(".c-product-block__metadata li").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Release date:") {
			if matches := yearRe.FindString(text); matches != "" {
				// Found year in release date metadata
				return
			}
		}
	})

	// Try meta tags as fallback
	doc.Find("meta[property='og:description'], meta[name='description']").Each(func(i int, s *goquery.Selection) {
		if content, exists := s.Attr("content"); exists {
			if matches := yearRe.FindString(content); matches != "" {
				// Found year in meta tag
				return
			}
		}
	})

	// Look for year in product information area as last resort
	yearText := doc.Find(".product-info").Text()
	if matches := yearRe.FindString(yearText); matches != "" {
		year := 0
		fmt.Sscanf(matches, "%d", &year)
		if year > 0 {
			return year, nil
		}
	}

	// Extract year from any matched text
	var foundYear int
	doc.Find(".c-product-block__metadata li").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Release date:") {
			if matches := yearRe.FindString(text); matches != "" {
				fmt.Sscanf(matches, "%d", &foundYear)
			}
		}
	})

	if foundYear > 0 {
		return foundYear, nil
	}

	return 0, fmt.Errorf("no year found")
}

// ParseCatalogAndLabel extracts both catalog number and label from product metadata.
func (p *PrestoParser) ParseCatalogAndLabel(html string) (catalog, label string, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Look for structured metadata in .c-product-block__metadata
	doc.Find(".c-product-block__metadata li").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		
		// Extract catalog number from "Catalogue number: XXX"
		if strings.Contains(text, "Catalogue number:") {
			catalog = strings.TrimPrefix(text, "Catalogue number:")
			catalog = strings.TrimSpace(catalog)
		}
		
		// Extract label from "Label: <a>XXX</a>"
		if strings.Contains(text, "Label:") {
			labelLink := s.Find("a")
			if labelLink.Length() > 0 {
				label = strings.TrimSpace(labelLink.Text())
			}
		}
	})

	if catalog == "" && label == "" {
		return "", "", fmt.Errorf("catalog and label not found in metadata")
	}

	return catalog, label, nil
}

// ParseCatalogNumber extracts just the catalog number (for compatibility).
func (p *PrestoParser) ParseCatalogNumber(html string) (string, error) {
	catalog, _, err := p.ParseCatalogAndLabel(html)
	return catalog, err
}

// ParseTracks extracts track listings from the semantic HTML structure.
func (p *PrestoParser) ParseTracks(html string) ([]TrackData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	tracks := make([]TrackData, 0)
	trackNum := 1

	// Parse using semantic structure: .c-tracklist__work elements
	doc.Find(".c-tracklist__work").Each(func(i int, work *goquery.Selection) {
		// Get composer and title using semantic links
		titleDiv := work.Find(".c-track__title").First()
		
		var composer, title string

		// Check if composer is in a link
		composerLink := titleDiv.Find("a[href*='composer']")
		if composerLink.Length() > 0 {
			composer = strings.TrimSpace(composerLink.Text())
			
			// Get work title from works link
			workLink := titleDiv.Find("a[href*='works']")
			if workLink.Length() > 0 {
				title = strings.TrimSpace(workLink.Text())
			}
		} else {
			// Composer might be plain text before a colon
			text := titleDiv.Text()
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				composer = strings.TrimSpace(parts[0])
				title = strings.TrimSpace(parts[1])
			} else {
				// No composer information, treat as anonymous
				composer = "Anonymous"
				title = strings.TrimSpace(text)
			}
		}

		// Clean up composer name
		composer = cleanHTMLEntities(composer)
		title = cleanHTMLEntities(title)

		if title != "" {
			track := TrackData{
				Disc:     1,
				Track:    trackNum,
				Title:    title,
				Composer: composer,
			}
			tracks = append(tracks, track)
			trackNum++
		}

		// Check for subtracks
		work.Find(".c-track--track").Each(func(j int, subtrack *goquery.Selection) {
			subtitleDiv := subtrack.Find(".c-track__title").First()
			subtitle := strings.TrimSpace(subtitleDiv.Text())
			subtitle = cleanHTMLEntities(subtitle)

			if subtitle != "" {
				track := TrackData{
					Disc:     1,
					Track:    trackNum,
					Title:    subtitle,
					Composer: composer, // Use parent work's composer
				}
				tracks = append(tracks, track)
				trackNum++
			}
		})
	})

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}