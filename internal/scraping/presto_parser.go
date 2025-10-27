package scraping

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cehbz/classical-tagger/internal/domain"
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
	data := &domain.Album{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]*domain.Track, 0),
	}

	result := &ExtractionResult{
		Album:  data,
		Source: "presto",
	}
	parsingNotes := make(map[string]interface{})

	// Parse title
	if title, err := p.ParseTitle(html); err == nil && title != "" {
		data.Title = title
	} else {
		result.Errors = append(result.Errors, ExtractionError{
			Field:    "title",
			Message:  "not found in HTML",
			Required: true,
		})
	}

	// Parse year from meta tag
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
	} else {
		result.Errors = append(result.Errors, ExtractionError{
			Field:    "year",
			Message:  "not found in HTML",
			Required: true,
		})
	}

	// Parse catalog number and label
	if catalog, label, err := p.ParseCatalogAndLabel(html); err == nil {
		edition := &domain.Edition{
			Label:         label,
			CatalogNumber: catalog,
			Year:          data.OriginalYear,
		}
		data.Edition = edition
	} else {
		result.Errors = append(result.Errors, ExtractionError{
			Field:    "catalog_number",
			Message:  "not found in HTML",
			Required: false,
		})
	}

	// Parse tracks using semantic structure
	if tracks, err := p.ParseTracks(html); err == nil && len(tracks) > 0 {
		data.Tracks = tracks
		parsingNotes["tracks_source"] = "semantic_structure"
	} else {
		result.Errors = append(result.Errors, ExtractionError{
			Field:    "tracks",
			Message:  "no tracks found in HTML",
			Required: true,
		})
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
	for key, value := range parsingNotes {
		result.Notes = append(result.Notes, fmt.Sprintf("%s: %v", key, value))
	}

	return result, nil
}

// ParseTitle extracts the album title from meta tags or HTML title tag.
func (p *PrestoParser) ParseTitle(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// PRIORITY 1: Check for og:title meta tag (most reliable)
	ogTitle, exists := doc.Find("meta[property='og:title']").Attr("content")
	if exists && ogTitle != "" {
		return cleanHTMLEntities(strings.TrimSpace(ogTitle)), nil
	}

	// PRIORITY 2: Check for h1.c-product-block__title
	h1Title := doc.Find("h1.c-product-block__title").First().Text()
	if h1Title != "" {
		return cleanHTMLEntities(strings.TrimSpace(h1Title)), nil
	}

	// PRIORITY 3: Fall back to parsing <title> tag
	title := doc.Find("title").First().Text()
	if title == "" {
		return "", fmt.Errorf("no title tag found")
	}

	// Remove " | Presto Music" suffix
	if idx := strings.Index(title, " | Presto Music"); idx > 0 {
		title = title[:idx]
	}

	// Remove everything after the FIRST " - " (label/catalog info)
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
func (p *PrestoParser) ParseTracks(html string) ([]*domain.Track, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	tracks := make([]*domain.Track, 0)
	trackNum := 1

	doc.Find(".c-tracklist__work").Each(func(i int, work *goquery.Selection) {
		// Check if this is a hierarchical work with movements
		hasChildren, parentTitle, parentComposer := p.detectHierarchy(work)

		if hasChildren {
			// Hierarchical structure: prepend parent to each movement
			work.Find(".c-track--track").Each(func(j int, subtrack *goquery.Selection) {
				subtitleDiv := subtrack.Find(".c-track__title").First()
				movementTitle := strings.TrimSpace(subtitleDiv.Text())
				movementTitle = cleanHTMLEntities(movementTitle)

				if movementTitle != "" {
					// Prepend parent work title to movement
					fullTitle := parentTitle + ": " + movementTitle

					track := &domain.Track{
						Disc:  1,
						Track: trackNum,
						Title: fullTitle,
						Artists: []domain.Artist{
							domain.Artist{
								Name: cleanHTMLEntities(parentComposer),
								Role: domain.RoleComposer,
							},
						},
					}
					tracks = append(tracks, track)
					trackNum++
				}
			})
		} else {
			// Flat structure: standard track
			titleDiv := work.Find(".c-track__title").First()

			var composer, title string

			composerLink := titleDiv.Find("a[href*='composer']")
			if composerLink.Length() > 0 {
				composer = strings.TrimSpace(composerLink.Text())

				workLink := titleDiv.Find("a[href*='works']")
				if workLink.Length() > 0 {
					title = strings.TrimSpace(workLink.Text())
				}
			} else {
				text := titleDiv.Text()
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					composer = strings.TrimSpace(parts[0])
					title = strings.TrimSpace(parts[1])
				} else {
					composer = "Anonymous"
					title = strings.TrimSpace(text)
				}
			}

			composer = cleanHTMLEntities(composer)
			title = cleanHTMLEntities(title)

			if title != "" {
				track := &domain.Track{
					Disc: 1,
					Track: trackNum,
					Title: title,
					Artists: []domain.Artist{
						domain.Artist{
							Name: composer,
							Role: domain.RoleComposer,
						},
					},
				}
				tracks = append(tracks, track)
				trackNum++
			}
		}
	})

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}

// detectHierarchy checks if a work element has subtrack children and extracts parent info.
func (p *PrestoParser) detectHierarchy(work *goquery.Selection) (hasChildren bool, parentTitle string, parentComposer string) {
	// Check for subtracks
	subtracks := work.Find(".c-track--track")
	if subtracks.Length() == 0 {
		return false, "", ""
	}

	// Get parent title and composer
	titleDiv := work.Find(".c-track__title").First()

	// Extract composer
	composerLink := titleDiv.Find("a[href*='composer']")
	if composerLink.Length() > 0 {
		parentComposer = strings.TrimSpace(composerLink.Text())
	} else {
		// Try plain text before colon
		text := titleDiv.Text()
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 2 {
			parentComposer = strings.TrimSpace(parts[0])
		}
	}

	// Extract work title
	workLink := titleDiv.Find("a[href*='works']")
	if workLink.Length() > 0 {
		parentTitle = strings.TrimSpace(workLink.Text())
	} else {
		// Try text after colon
		text := titleDiv.Text()
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 2 {
			parentTitle = strings.TrimSpace(parts[1])
		}
	}

	return true, parentTitle, parentComposer
}
