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

	// Parse album-level performers and set as album artist
	performers, dups, err := p.ParsePerformers(html)
	albumLevelPerformers := performers
	if err == nil && len(performers) > 0 {
		parsingNotes["album_performers_found"] = len(performers)

		// Add deduplication notes for transparency
		if len(dups) > 0 {
			parsingNotes["dups"] = dups
			// Also add to warnings so they're visible by default
			for _, note := range dups {
				result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ %s", note))
			}
		}

		// Format performers as album artist (don't merge into tracks)
		data.AlbumArtist = performers
	} else if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("album performers: %v", err))
	}

	// After parsing tracks and album-level performers, check for universal performers in tracks
	// and merge with album-level performers if any exist. Do NOT remove from tracks; instead ensure inclusion.
	if len(data.Tracks) > 0 {
		universalTrackArtists := domain.DetermineAlbumArtistFromAlbum(data)

		if len(universalTrackArtists) > 0 {
			if len(albumLevelPerformers) > 0 {
				// Merge album-level and track-derived performers (avoid duplicates)
				combinedPerformers := mergePerformers(albumLevelPerformers, universalTrackArtists)
				data.AlbumArtist = combinedPerformers
			} else {
				// No album-level performers, but found universal performers in tracks
				data.AlbumArtist = universalTrackArtists
			}
		} else if len(albumLevelPerformers) > 0 {
			// Only album-level performers, no universal performers in tracks
			data.AlbumArtist = albumLevelPerformers
		}

		// Ensure AlbumArtist performers appear on each track unless it's Various Artists
		if len(data.AlbumArtist) > 0 {
			if !strings.EqualFold(strings.TrimSpace(domain.FormatArtists(data.AlbumArtist)), "Various Artists") {
				ensureArtistsOnTracks(data.Tracks, data.AlbumArtist)
			}
		}
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

	// Convert Album to Torrent before returning (use empty root path for web scrapers)
	result.Torrent = data.ToTorrent("")

	return result, nil
}

// ParsePerformers extracts album-level performers from page GraphQL data.
// Returns a list of performers (ensemble, conductor, soloists) that appear on the album level.
// Prioritizes releaseCredits from GraphQL data which includes role information.
// Also returns deduplication information for transparency.
func (p *DiscogsParser) ParsePerformers(html string) ([]domain.Artist, []string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Step 1: Get performer names from JSON-LD byArtist (authoritative list)
	jsonLD := doc.Find("script#release_schema[type='application/ld+json']").First().Text()
	if jsonLD == "" {
		return nil, nil, fmt.Errorf("no JSON-LD release_schema found")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonLD), &data); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	performerNames := make([]string, 0)

	if releaseOf, ok := data["releaseOf"].(map[string]interface{}); ok {
		if byArtist, ok := releaseOf["byArtist"].([]interface{}); ok {
			for _, artistData := range byArtist {
				if artistMap, ok := artistData.(map[string]interface{}); ok {
					if name, ok := artistMap["name"].(string); ok && name != "" {
						performerNames = append(performerNames, strings.TrimSpace(name))
					}
				}
			}
		}
	}

	if len(performerNames) == 0 {
		return nil, nil, fmt.Errorf("no performers found in JSON-LD byArtist")
	}

	// Step 2: Extract role mappings from releaseCredits in Apollo state
	roleMap, _ := p.extractReleaseCredits(html)
	// If GraphQL extraction fails or is incomplete, attempt to parse HTML Credits section
	if len(roleMap) == 0 {
		if htmlRoles, err := p.parseCreditsFromHTML(html); err == nil && len(htmlRoles) > 0 {
			roleMap = htmlRoles
		}
	}
	// If still empty, we'll infer roles as a last resort below

	// Step 3: Build performer list by matching names with roles
	performers := make([]domain.Artist, 0)
	seen := make(map[string]bool)                   // Exact name tracking
	normalizedSeen := make(map[string]bool)         // Normalized name tracking for deduplication
	normalizedToOriginal := make(map[string]string) // Track which original name we kept for each normalized form
	dups := make([]string, 0)                       // Track deduplications for transparency

	for _, name := range performerNames {
		normalized := normalizeNameForDedup(name)
		if normalizedSeen[normalized] {
			// Deduplication occurred - record it
			original := normalizedToOriginal[normalized]
			if original != name {
				dups = append(dups, fmt.Sprintf("duplication: '%s' merged with '%s' (normalized: '%s')", name, original, normalized))
			}
			continue
		}
		seen[name] = true
		normalizedSeen[normalized] = true
		normalizedToOriginal[normalized] = name // Keep track of which original name we're using

		var role domain.Role

		// Try to get role from releaseCredits (check both exact name and normalized)
		roleStr := ""
		if r, hasRole := roleMap[name]; hasRole {
			roleStr = r
		} else {
			// Try to find by normalized name
			for k, v := range roleMap {
				if normalizeNameForDedup(k) == normalized {
					roleStr = v
					break
				}
			}
		}

		if roleStr != "" {
			// Try deterministic mapping
			if mappedRole, ok := mapDiscogsRoleToDomainRole(roleStr); ok {
				role = mappedRole
			} else {
				// Role string exists but not mappable, infer from name
				role = inferRoleFromName(name)
			}
		} else {
			// No role in releaseCredits, must infer from name
			role = inferRoleFromName(name)
		}

		performers = append(performers, domain.Artist{
			Name: name,
			Role: role,
		})
	}

	// Step 4: Add any performers from releaseCredits that weren't in JSON-LD
	for name, roleStr := range roleMap {
		normalized := normalizeNameForDedup(name)
		if normalizedSeen[normalized] {
			// Deduplication occurred - record it
			original := normalizedToOriginal[normalized]
			if original != name {
				dups = append(dups, fmt.Sprintf("duplicate: '%s' merged with '%s' (normalized: '%s')", name, original, normalized))
			}
			continue
		}
		seen[name] = true
		normalizedSeen[normalized] = true
		normalizedToOriginal[normalized] = name

		var role domain.Role
		if mappedRole, ok := mapDiscogsRoleToDomainRole(roleStr); ok {
			role = mappedRole
		} else {
			role = inferRoleFromName(name)
		}

		performers = append(performers, domain.Artist{
			Name: name,
			Role: role,
		})
	}

	if len(performers) == 0 {
		return nil, nil, fmt.Errorf("no performers extracted")
	}

	return performers, dups, nil
}

// parseCreditsFromHTML parses the Discogs HTML "Credits" section to a name->role map.
// It looks for common credits blocks/tables and extracts display name and credit role text.
func (p *DiscogsParser) parseCreditsFromHTML(html string) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML for credits: %w", err)
	}

	roles := make(map[string]string)

	// New Discogs layout: credits are often under a section with headings like "Credits"
	// and list items with role and name links. We try a few common selectors.
	// 1) Credits list items
	doc.Find("section:contains('Credits') li").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}
		// Try to split "Role – Name" or "Role: Name"
		parts := strings.FieldsFunc(text, func(r rune) bool { return r == '–' || r == '-' || r == ':' })
		if len(parts) >= 2 {
			role := strings.TrimSpace(parts[0])
			name := strings.TrimSpace(strings.Join(parts[1:], " "))
			if name != "" && role != "" {
				roles[name] = role
			}
		}
	})

	// 2) Table rows that look like credits (fallback heuristic)
	if len(roles) == 0 {
		doc.Find("table tr").Each(func(i int, tr *goquery.Selection) {
			tds := tr.Find("td")
			if tds.Length() >= 2 {
				role := strings.TrimSpace(tds.Eq(0).Text())
				name := strings.TrimSpace(tds.Eq(1).Text())
				if role != "" && name != "" {
					roles[name] = role
				}
			}
		})
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no credits found in HTML")
	}
	return roles, nil
}

// extractReleaseCredits extracts the releaseCredits from Apollo GraphQL state.
// Returns a map of artist name -> role string.
func (p *DiscogsParser) extractReleaseCredits(html string) (map[string]string, error) {
	// Find the releaseCredits array in Apollo state
	startIdx := strings.Index(html, `"releaseCredits":[`)
	if startIdx == -1 {
		return nil, fmt.Errorf("no releaseCredits found")
	}

	// Find matching closing bracket for the array
	arrayStart := startIdx + len(`"releaseCredits":`)
	depth := 0
	inString := false
	escape := false
	endIdx := -1

	for i := arrayStart; i < len(html) && i < arrayStart+50000; i++ {
		ch := html[i]

		if escape {
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if ch == '[' || ch == '{' {
			depth++
		} else if ch == ']' || ch == '}' {
			depth--
			if depth == 0 && ch == ']' {
				endIdx = i + 1
				break
			}
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("could not find end of releaseCredits array")
	}

	creditsJSON := html[arrayStart:endIdx]

	var credits []map[string]interface{}
	if err := json.Unmarshal([]byte(creditsJSON), &credits); err != nil {
		return nil, fmt.Errorf("failed to parse releaseCredits: %w", err)
	}

	roleMap := make(map[string]string)

	for _, credit := range credits {
		displayName, _ := credit["displayName"].(string)
		creditRole, _ := credit["creditRole"].(string)

		if displayName != "" {
			displayName = strings.TrimSpace(displayName)
			roleMap[displayName] = creditRole
		}

		// Also check nameVariation field (sometimes used)
		if nameVar, ok := credit["nameVariation"].(string); ok && nameVar != "" {
			nameVar = strings.TrimSpace(nameVar)
			// Add variation as alternative key for matching
			if _, exists := roleMap[nameVar]; !exists {
				roleMap[nameVar] = creditRole
			}
		}
	}

	return roleMap, nil
}

// normalizeNameForDedup normalizes artist names for deduplication purposes.
// Aggressively normalizes by keeping only letters and numbers (removes all punctuation and spaces).
// This handles variations like "RIAS-Kammerchor" vs "RIAS Kammerchor" vs "RIASKammerchor".
// Numbers are kept to distinguish different ensembles (e.g., "Orchestra 1" vs "Orchestra 2").
func normalizeNameForDedup(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	// Remove everything except letters and numbers
	var result strings.Builder
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// mapDiscogsRoleToDomainRole maps Discogs role strings to domain.Role.
// Returns (role, true) if deterministically mappable, (empty, false) otherwise.
func mapDiscogsRoleToDomainRole(discogsRole string) (domain.Role, bool) {
	roleLower := strings.ToLower(strings.TrimSpace(discogsRole))

	// Ensemble/Orchestra indicators
	ensembleRoles := map[string]bool{
		"choir":          true,
		"chorus":         true,
		"orchestra":      true,
		"ensemble":       true,
		"vocal ensemble": true,
		"chamber choir":  true,
		"kammerchor":     true,
	}

	if ensembleRoles[roleLower] {
		return domain.RoleEnsemble, true
	}

	// Conductor indicators
	conductorRoles := map[string]bool{
		"conductor":     true,
		"chorus master": true,
		"chorusmaster":  true,
		"director":      true,
		"maestro":       true,
	}

	if conductorRoles[roleLower] {
		return domain.RoleConductor, true
	}

	// Soloist indicators
	soloistRoles := map[string]bool{
		"soloist":         true,
		"vocalist":        true,
		"singer":          true,
		"performer":       true,
		"instrumentalist": true,
	}

	if soloistRoles[roleLower] {
		return domain.RoleSoloist, true
	}

	// Cannot deterministically map
	return domain.RoleUnknown, false
}

// inferRoleFromName attempts to infer artist role from their name.
// This is a fallback heuristic when role string is unavailable.
func inferRoleFromName(name string) domain.Role {
	nameLower := strings.ToLower(name)

	// Check for explicit role indicators in name
	if strings.Contains(nameLower, "conductor") || strings.Contains(nameLower, "director") {
		return domain.RoleConductor
	}

	// Check for ensemble indicators
	ensembleKeywords := []string{
		"orchestra", "philharmonic", "symphony", "ensemble",
		"choir", "chorus", "kammerchor", "kammer",
		"quartet", "trio", "quintet", "sextet",
		"chamber", "band", "consort", "players",
	}

	for _, keyword := range ensembleKeywords {
		if strings.Contains(nameLower, keyword) {
			return domain.RoleEnsemble
		}
	}

	// Default to soloist for individual names
	return domain.RoleSoloist
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
				Disc:  1,
				Track: trackNum,
				Title: title,
			}
			if composer != "" {
				track.Artists = append(track.Artists, domain.Artist{Name: composer, Role: domain.RoleComposer})
			}
			tracks = append(tracks, track)
		}
	})

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	return tracks, nil
}
