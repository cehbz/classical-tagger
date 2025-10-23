package scraping

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// PrestoClassicalParser parses Presto Classical HTML pages.
// Presto Classical is a UK-based classical music retailer with excellent metadata.
type PrestoClassicalParser struct {
	// Parser is stateless and immutable
}

// NewPrestoClassicalParser creates a new parser instance.
func NewPrestoClassicalParser() *PrestoClassicalParser {
	return &PrestoClassicalParser{}
}

// Parse parses a complete Presto Classical HTML page and returns extraction result.
func (p *PrestoClassicalParser) Parse(html string) (*ExtractionResult, error) {
	data := &AlbumData{
		Title:        MissingTitle,
		OriginalYear: MissingYear,
		Tracks:       make([]TrackData, 0),
	}

	result := NewExtractionResult(data)
	parsingNotes := make(map[string]interface{})

	// Parse JSON-LD for basic metadata
	if err := p.parseJSONLD(html, data, result, parsingNotes); err != nil {
		result = result.WithWarning(fmt.Sprintf("JSON-LD parsing: %v", err))
	}

	// Parse title from H1 if not from JSON-LD
	if data.Title == MissingTitle {
		if title, err := p.ParseTitle(html); err == nil && title != "" {
			data.Title = title
		} else {
			result = result.WithError(NewExtractionError("title", "not found in HTML", true))
		}
	}

	// Parse tracks from HTML
	tracks, trackErrors := p.ParseTracks(html)
	if len(tracks) > 0 {
		data.Tracks = tracks
		parsingNotes["track_count"] = len(tracks)
	} else {
		result = result.WithError(NewExtractionError("tracks", "no tracks found", true))
	}

	// Add any track parsing errors
	for _, err := range trackErrors {
		result = result.WithWarning(err.Error())
	}

	// Parse year - try multiple methods
	if year, err := p.ParseYear(html); err == nil && year > 0 {
		data.OriginalYear = year
		parsingNotes["year_source"] = "page_content"
	} else {
		// Year is often missing from Presto - mark as warning not error
		result = result.WithWarning("Recording year not found - may need manual entry")
		parsingNotes["year_source"] = "not_found"
	}

	// Parse edition information
	if edition, err := p.ParseEdition(html); err == nil {
		data.Edition = edition
		parsingNotes["edition_parsed"] = true
	}

	result = result.WithParsingNotes(parsingNotes)
	return result, nil
}

// ParseTitle extracts album title from H1 element.
func (p *PrestoClassicalParser) ParseTitle(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Title is in H1 tag
	title := doc.Find("h1").First().Text()
	title = strings.TrimSpace(title)

	// Remove " (Digital Download)" suffix if present
	title = strings.TrimSuffix(title, " (Digital Download)")
	title = strings.TrimSpace(title)

	if title == "" {
		return "", fmt.Errorf("title not found")
	}

	return decodeHTMLEntities(title), nil
}

// ParseYear attempts to extract recording year from the page.
// Note: Presto Classical often doesn't display recording year prominently.
func (p *PrestoClassicalParser) ParseYear(html string) (int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return 0, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Try to find year in various places
	yearPattern := regexp.MustCompile(`\b(19|20)\d{2}\b`)

	// Check album info sections
	infoText := doc.Find(".c-product__info").Text()
	if matches := yearPattern.FindStringSubmatch(infoText); len(matches) > 0 {
		year, _ := strconv.Atoi(matches[0])
		if year >= 1900 && year <= 2030 {
			return year, nil
		}
	}

	// Check product details
	detailsText := doc.Find(".c-product-details").Text()
	if matches := yearPattern.FindStringSubmatch(detailsText); len(matches) > 0 {
		year, _ := strconv.Atoi(matches[0])
		if year >= 1900 && year <= 2030 {
			return year, nil
		}
	}

	return 0, fmt.Errorf("year not found")
}

// ParseTracks extracts track listing from the tracklist section.
func (p *PrestoClassicalParser) ParseTracks(html string) ([]TrackData, []error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, []error{fmt.Errorf("failed to parse HTML: %w", err)}
	}

	tracks := make([]TrackData, 0)
	errors := make([]error, 0)
	trackNum := 1

	// Find the main tracklist container
	tracklist := doc.Find(".c-tracklist")
	if tracklist.Length() == 0 {
		return nil, []error{fmt.Errorf("tracklist container not found")}
	}

	// Find all works
	works := tracklist.Find(".c-tracklist__work")
	works.Each(func(workIdx int, work *goquery.Selection) {
		// Parse the work-level track (composer + work title)
		workTrack := work.Find(".c-track--work").First()
		
		var composer string
		var workTitle string

		// Extract composer
		composerLink := workTrack.Find("a[href*='composer']").First()
		if composerLink.Length() > 0 {
			composer = strings.TrimSpace(composerLink.Text())
			// Format: "Bach, J S" -> "Johann Sebastian Bach" (keep as is for now)
		}

		// Extract work title
		workTitleLink := workTrack.Find("a.c-track__title").First()
		if workTitleLink.Length() > 0 {
			workTitle = strings.TrimSpace(workTitleLink.Text())
		}

		// Find all individual movement tracks
		movementTracks := work.Find(".c-track--track")
		
		if movementTracks.Length() == 0 {
			// Single-movement work - use work title as track
			track := TrackData{
				Disc:     1, // Default to disc 1
				Track:    trackNum,
				Title:    workTitle,
				Composer: composer,
				Artists:  make([]ArtistData, 0),
			}
			tracks = append(tracks, track)
			trackNum++
		} else {
			// Multi-movement work
			movementTracks.Each(func(mvtIdx int, mvtTrack *goquery.Selection) {
				// Extract movement title
				titleSpan := mvtTrack.Find("span.c-track__title").First()
				movementTitle := strings.TrimSpace(titleSpan.Text())

				// If movement title is empty, use work title
				if movementTitle == "" {
					movementTitle = workTitle
				}

				// Extract duration
				durationElem := mvtTrack.Find(".c-track__duration").First()
				durationText := strings.TrimSpace(durationElem.Text())
				// Format: "Track length2:33" -> "2:33"
				durationText = strings.TrimPrefix(durationText, "Track length")
				durationText = strings.TrimSpace(durationText)

				// Extract performers
				performersElem := mvtTrack.Find(".c-track__performers").First()
				performersText := strings.TrimSpace(performersElem.Text())
				
				artists := make([]ArtistData, 0)
				if performersText != "" {
					// Parse performers - they're usually comma-separated
					performers := strings.Split(performersText, ",")
					for _, performer := range performers {
						performer = strings.TrimSpace(performer)
						if performer != "" {
							// Try to infer role (basic heuristic)
							role := inferArtistRole(performer, movementTitle)
							artists = append(artists, ArtistData{
								Name: performer,
								Role: role,
							})
						}
					}
				}

				track := TrackData{
					Disc:     1, // Default to disc 1
					Track:    trackNum,
					Title:    movementTitle,
					Composer: composer,
					Artists:  artists,
				}

				tracks = append(tracks, track)
				trackNum++
			})
		}
	})

	if len(tracks) == 0 {
		errors = append(errors, fmt.Errorf("no tracks parsed from HTML"))
	}

	return tracks, errors
}

// ParseEdition extracts label and catalog number from JSON-LD or page.
func (p *PrestoClassicalParser) ParseEdition(html string) (*EditionData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	edition := &EditionData{}

	// Try to extract from JSON-LD first
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonText := s.Text()
		
		// Look for MPN (Manufacturer Part Number) - this is often the catalog number
		mpnPattern := regexp.MustCompile(`"mpn"\s*:\s*"([^"]+)"`)
		if matches := mpnPattern.FindStringSubmatch(jsonText); len(matches) > 1 {
			edition.CatalogNumber = matches[1]
		}

		// Look for brand/label
		brandPattern := regexp.MustCompile(`"brand"\s*:\s*\{\s*"name"\s*:\s*"([^"]+)"`)
		if matches := brandPattern.FindStringSubmatch(jsonText); len(matches) > 1 {
			edition.Label = matches[1]
		}
	})

	// If we found catalog number but no label, try to extract label from product details
	if edition.CatalogNumber != "" && edition.Label == "" {
		// Look in product info area
		productInfo := doc.Find(".c-product__info, .c-product-details").Text()
		
		// Common label patterns
		labelPattern := regexp.MustCompile(`(?i)Label:\s*([^\n]+)`)
		if matches := labelPattern.FindStringSubmatch(productInfo); len(matches) > 1 {
			edition.Label = strings.TrimSpace(matches[1])
		}
	}

	if edition.CatalogNumber == "" && edition.Label == "" {
		return nil, fmt.Errorf("no edition information found")
	}

	return edition, nil
}

// parseJSONLD attempts to extract metadata from JSON-LD structured data.
func (p *PrestoClassicalParser) parseJSONLD(html string, data *AlbumData, result *ExtractionResult, notes map[string]interface{}) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find JSON-LD scripts
	jsonldFound := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonText := s.Text()
		
		// Extract product name
		namePattern := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
		if matches := namePattern.FindStringSubmatch(jsonText); len(matches) > 1 {
			name := matches[1]
			// Remove format suffix like "(Digital Download)"
			name = regexp.MustCompile(`\s*\([^)]*Download[^)]*\)`).ReplaceAllString(name, "")
			name = strings.TrimSpace(name)
			
			if name != "" && data.Title == MissingTitle {
				data.Title = decodeHTMLEntities(name)
				jsonldFound = true
			}
		}

		// Extract MPN as catalog number
		mpnPattern := regexp.MustCompile(`"mpn"\s*:\s*"([^"]+)"`)
		if matches := mpnPattern.FindStringSubmatch(jsonText); len(matches) > 1 {
			if data.Edition == nil {
				data.Edition = &EditionData{}
			}
			data.Edition.CatalogNumber = matches[1]
		}

		// Extract GTIN (UPC/EAN)
		gtinPattern := regexp.MustCompile(`"gtin13"\s*:\s*"([^"]+)"`)
		if matches := gtinPattern.FindStringSubmatch(jsonText); len(matches) > 1 {
			notes["gtin"] = matches[1]
		}
	})

	if !jsonldFound {
		return fmt.Errorf("no JSON-LD data found")
	}

	notes["jsonld_parsed"] = true
	return nil
}

// inferArtistRole attempts to determine the artist's role based on context.
// This is a basic heuristic and may need refinement.
func inferArtistRole(name, context string) string {
	nameLower := strings.ToLower(name)
	contextLower := strings.ToLower(context)

	// Look for role indicators in the name itself
	if strings.Contains(nameLower, "conductor") {
		return "conductor"
	}
	if strings.Contains(nameLower, "orchestra") || strings.Contains(nameLower, "ensemble") ||
		strings.Contains(nameLower, "quartet") || strings.Contains(nameLower, "choir") {
		return "ensemble"
	}

	// Look for instruments in context
	instruments := []string{"piano", "violin", "cello", "flute", "clarinet", "trumpet", "horn"}
	for _, inst := range instruments {
		if strings.Contains(contextLower, inst) {
			return "soloist"
		}
	}

	// Default to performer
	return "performer"
}