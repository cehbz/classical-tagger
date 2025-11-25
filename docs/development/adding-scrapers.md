# Adding Metadata Scrapers

This guide shows you how to add support for new classical music metadata sources.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Step-by-Step Guide](#step-by-step-guide)
- [Implementation Pattern](#implementation-pattern)
- [Testing](#testing)
- [Integration](#integration)
- [Best Practices](#best-practices)

---

## Overview

A metadata scraper (extractor) fetches classical music metadata from a website and converts it to our domain model. Each scraper implements the `Extractor` interface and handles a specific website.

**Typical workflow:**
1. Fetch HTML from URL
2. Parse HTML using goquery (CSS selectors)
3. Extract metadata fields
4. Convert to domain model
5. Return structured result

**Supported sources:** See [Metadata Sources Reference](../development/metadata-sources.md)

---

## Prerequisites

### Required Knowledge

- Go programming (basic)
- HTML and CSS selectors
- HTTP requests
- Classical music metadata structure

### Required Tools

```bash
# goquery for HTML parsing
go get github.com/PuerkitoBio/goquery

# HTTP client (standard library)
# Already available in net/http
```

### Study the Target Website

Before coding, understand the website structure:

```bash
# 1. Fetch sample page
curl -s "https://example.com/album/123" > test.html

# 2. Open in browser with DevTools
# Inspect HTML structure
# Note CSS selectors for each field

# 3. Check robots.txt
curl "https://example.com/robots.txt"

# 4. Check rate limits
# Look for rate limit info in API docs or terms of service
```

---

## Step-by-Step Guide

### Step 1: Create Extractor File

Create a new file: `internal/scraping/sitename_extractor.go`

```go
package scraping

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cehbz/classical-tagger/internal/domain"
)

// SiteNameExtractor extracts metadata from sitename.com
// Example URLs:
//   - https://sitename.com/album/12345
//   - https://sitename.com/release/some-album-title
type SiteNameExtractor struct{}

// NewSiteNameExtractor creates a new SiteName extractor.
func NewSiteNameExtractor() *SiteNameExtractor {
	return &SiteNameExtractor{}
}

// CanHandle returns true if this extractor can handle the given URL.
func (e *SiteNameExtractor) CanHandle(url string) bool {
	return strings.Contains(url, "sitename.com")
}

// Extract fetches and parses metadata from the URL.
func (e *SiteNameExtractor) Extract(url string) (*ExtractionResult, error) {
	// Implementation goes here
	return nil, fmt.Errorf("not implemented")
}
```

---

### Step 2: Implement URL Detection

Make `CanHandle()` precise:

```go
func (e *SiteNameExtractor) CanHandle(url string) bool {
	// Match specific patterns
	return strings.Contains(url, "sitename.com/album/") ||
	       strings.Contains(url, "sitename.com/release/")
}
```

**Tips:**
- Be specific to avoid false positives
- Handle both HTTP and HTTPS
- Support multiple URL patterns if needed
- Don't match unrelated pages (homepage, search, etc.)

---

### Step 3: Implement HTML Fetching

```go
func (e *SiteNameExtractor) Extract(url string) (*ExtractionResult, error) {
	// Fetch HTML
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract metadata
	return e.parseDocument(doc, url)
}
```

---

### Step 4: Implement Metadata Parsing

Create a `parseDocument()` method:

```go
func (e *SiteNameExtractor) parseDocument(doc *goquery.Document, sourceURL string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Source:    "SiteName",
		SourceURL: sourceURL,
		Torrent:   &domain.Torrent{},
	}

	// Extract album title
	title := doc.Find(".album-title").First().Text()
	title = strings.TrimSpace(title)
	if title == "" {
		result.AddError("title", "Album title not found")
	} else {
		result.Torrent.Title = title
	}

	// Extract year
	yearStr := doc.Find(".release-year").First().Text()
	yearStr = strings.TrimSpace(yearStr)
	if year, err := strconv.Atoi(yearStr); err == nil {
		result.Torrent.OriginalYear = year
	} else {
		result.AddWarning("year", "Release year not found or invalid")
	}

	// Extract edition info
	label := doc.Find(".record-label").First().Text()
	catalogNum := doc.Find(".catalog-number").First().Text()
	if label != "" && catalogNum != "" {
		result.Torrent.Edition = &domain.Edition{
			Label:         strings.TrimSpace(label),
			CatalogNumber: strings.TrimSpace(catalogNum),
			Year:          result.Torrent.OriginalYear,
		}
	}

	// Extract tracks
	tracks, err := e.extractTracks(doc)
	if err != nil {
		result.AddError("tracks", fmt.Sprintf("Failed to extract tracks: %v", err))
	} else if len(tracks) == 0 {
		result.AddError("tracks", "No tracks found")
	} else {
		result.Torrent.Files = tracks
	}

	return result, nil
}
```

---

### Step 5: Implement Track Extraction

```go
func (e *SiteNameExtractor) extractTracks(doc *goquery.Document) ([]domain.FileLike, error) {
	var files []domain.FileLike
	disc := 1
	trackNum := 1

	doc.Find(".track-list .track").Each(func(i int, s *goquery.Selection) {
		// Check for disc header
		if discHeader := s.Find(".disc-header").First(); discHeader.Length() > 0 {
			if newDisc, err := strconv.Atoi(discHeader.Text()); err == nil {
				disc = newDisc
				trackNum = 1
				return // Skip this iteration
			}
		}

		// Extract track info
		title := s.Find(".track-title").First().Text()
		title = strings.TrimSpace(title)
		if title == "" {
			return // Skip empty tracks
		}

		// Extract composer
		composer := s.Find(".composer").First().Text()
		composer = strings.TrimSpace(composer)

		// Extract performers
		performer := s.Find(".performer").First().Text()
		performer = strings.TrimSpace(performer)

		// Build artists list
		var artists []domain.Artist
		if composer != "" {
			artists = append(artists, domain.Artist{
				Name: composer,
				Role: domain.RoleComposer,
			})
		}
		if performer != "" {
			// Infer role from performer string
			role := inferPerformerRole(performer)
			artists = append(artists, domain.Artist{
				Name: performer,
				Role: role,
			})
		}

		// Create track
		track := &domain.Track{
			File: domain.File{
				Path: fmt.Sprintf("%02d - %s.flac", trackNum, sanitizeFilename(title)),
			},
			Disc:    disc,
			Track:   trackNum,
			Title:   title,
			Artists: artists,
		}

		files = append(files, track)
		trackNum++
	})

	return files, nil
}
```

---

### Step 6: Add Helper Functions

```go
// inferPerformerRole attempts to determine the performer's role.
func inferPerformerRole(performer string) domain.Role {
	lower := strings.ToLower(performer)
	
	// Check for keywords
	if strings.Contains(lower, "piano") || strings.Contains(lower, "violin") {
		return domain.RoleSoloist
	}
	if strings.Contains(lower, "orchestra") || strings.Contains(lower, "ensemble") {
		return domain.RoleEnsemble
	}
	if strings.Contains(lower, "conductor") {
		return domain.RoleConductor
	}
	
	// Default to soloist
	return domain.RoleSoloist
}

// sanitizeFilename removes invalid characters from a filename.
func sanitizeFilename(s string) string {
	// Remove invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		s = strings.ReplaceAll(s, char, "")
	}
	
	// Trim spaces
	s = strings.TrimSpace(s)
	
	return s
}
```

---

### Step 7: Handle Edge Cases

```go
func (e *SiteNameExtractor) parseDocument(doc *goquery.Document, sourceURL string) (*ExtractionResult, error) {
	// ... existing code ...

	// Handle multi-disc albums
	if doc.Find(".disc-header").Length() > 1 {
		result.AddNote("Multi-disc album detected")
	}

	// Handle compilation albums
	if strings.Contains(strings.ToLower(result.Torrent.Title), "various artists") {
		result.AddWarning("album", "Compilation album detected - verify artists")
	}

	// Handle missing data
	if result.Torrent.Edition == nil {
		result.AddWarning("edition", "Edition information not found")
	}

	// Validate extracted data
	if result.Torrent.Title == "" {
		result.AddError("title", "Required field: title is missing")
	}
	if result.Torrent.OriginalYear == 0 {
		result.AddError("year", "Required field: year is missing")
	}
	if len(result.Torrent.Files) == 0 {
		result.AddError("tracks", "Required field: no tracks found")
	}

	return result, nil
}
```

---

## Implementation Pattern

### Complete Example Structure

```go
package scraping

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cehbz/classical-tagger/internal/domain"
)

// SiteNameExtractor extracts metadata from sitename.com
type SiteNameExtractor struct{}

func NewSiteNameExtractor() *SiteNameExtractor {
	return &SiteNameExtractor{}
}

func (e *SiteNameExtractor) CanHandle(url string) bool {
	return strings.Contains(url, "sitename.com/album/")
}

func (e *SiteNameExtractor) Extract(url string) (*ExtractionResult, error) {
	// 1. Fetch HTML
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 2. Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	// 3. Extract metadata
	return e.parseDocument(doc, url)
}

func (e *SiteNameExtractor) parseDocument(doc *goquery.Document, sourceURL string) (*ExtractionResult, error) {
	result := &ExtractionResult{
		Source:    "SiteName",
		SourceURL: sourceURL,
		Torrent:   &domain.Torrent{},
	}

	// Extract required fields
	result.Torrent.Title = e.extractTitle(doc, result)
	result.Torrent.OriginalYear = e.extractYear(doc, result)
	result.Torrent.Edition = e.extractEdition(doc, result)
	
	// Extract tracks
	tracks, err := e.extractTracks(doc)
	if err != nil {
		result.AddError("tracks", err.Error())
	} else {
		result.Torrent.Files = tracks
	}

	return result, nil
}

func (e *SiteNameExtractor) extractTitle(doc *goquery.Document, result *ExtractionResult) string {
	title := doc.Find(".album-title").First().Text()
	title = strings.TrimSpace(title)
	if title == "" {
		result.AddError("title", "Album title not found")
	}
	return title
}

func (e *SiteNameExtractor) extractYear(doc *goquery.Document, result *ExtractionResult) int {
	yearStr := doc.Find(".release-year").First().Text()
	year, err := strconv.Atoi(strings.TrimSpace(yearStr))
	if err != nil {
		result.AddWarning("year", "Year not found or invalid")
		return 0
	}
	return year
}

func (e *SiteNameExtractor) extractEdition(doc *goquery.Document, result *ExtractionResult) *domain.Edition {
	label := strings.TrimSpace(doc.Find(".label").First().Text())
	catalog := strings.TrimSpace(doc.Find(".catalog").First().Text())
	
	if label == "" || catalog == "" {
		result.AddWarning("edition", "Edition info incomplete")
		return nil
	}
	
	return &domain.Edition{
		Label:         label,
		CatalogNumber: catalog,
		Year:          result.Torrent.OriginalYear,
	}
}

func (e *SiteNameExtractor) extractTracks(doc *goquery.Document) ([]domain.FileLike, error) {
	var files []domain.FileLike
	disc := 1
	trackNum := 1

	doc.Find(".track").Each(func(i int, s *goquery.Selection) {
		// Extract and create track
		track := e.parseTrackElement(s, disc, trackNum)
		if track != nil {
			files = append(files, track)
			trackNum++
		}
	})

	return files, nil
}

func (e *SiteNameExtractor) parseTrackElement(s *goquery.Selection, disc, trackNum int) *domain.Track {
	title := strings.TrimSpace(s.Find(".track-title").Text())
	if title == "" {
		return nil
	}

	composer := strings.TrimSpace(s.Find(".composer").Text())
	performer := strings.TrimSpace(s.Find(".performer").Text())

	var artists []domain.Artist
	if composer != "" {
		artists = append(artists, domain.Artist{
			Name: composer,
			Role: domain.RoleComposer,
		})
	}
	if performer != "" {
		artists = append(artists, domain.Artist{
			Name: performer,
			Role: inferRole(performer),
		})
	}

	return &domain.Track{
		File: domain.File{
			Path: fmt.Sprintf("%02d - %s.flac", trackNum, sanitizeFilename(title)),
		},
		Disc:    disc,
		Track:   trackNum,
		Title:   title,
		Artists: artists,
	}
}

func inferRole(performer string) domain.Role {
	lower := strings.ToLower(performer)
	if strings.Contains(lower, "conductor") {
		return domain.RoleConductor
	}
	if strings.Contains(lower, "orchestra") || strings.Contains(lower, "ensemble") {
		return domain.RoleEnsemble
	}
	return domain.RoleSoloist
}

func sanitizeFilename(s string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		s = strings.ReplaceAll(s, char, "")
	}
	return strings.TrimSpace(s)
}
```

---

## Testing

### Step 1: Create Test File

Create `internal/scraping/sitename_extractor_test.go`:

```go
package scraping

import (
	"testing"
)

func TestSiteNameExtractor_CanHandle(t *testing.T) {
	e := NewSiteNameExtractor()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid album URL",
			url:  "https://sitename.com/album/12345",
			want: true,
		},
		{
			name: "valid release URL",
			url:  "https://sitename.com/release/some-album",
			want: true,
		},
		{
			name: "different site",
			url:  "https://othersite.com/album/12345",
			want: false,
		},
		{
			name: "homepage",
			url:  "https://sitename.com/",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.CanHandle(tt.url)
			if got != tt.want {
				t.Errorf("CanHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSiteNameExtractor_parseDocument(t *testing.T) {
	// Mock HTML for testing
	html := `
		<html>
			<div class="album-title">Goldberg Variations</div>
			<div class="release-year">1981</div>
			<div class="label">Sony Classical</div>
			<div class="catalog">SMK89245</div>
			<div class="track-list">
				<div class="track">
					<div class="track-title">Aria</div>
					<div class="composer">J.S. Bach</div>
					<div class="performer">Glenn Gould (piano)</div>
				</div>
			</div>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	e := NewSiteNameExtractor()
	result, err := e.parseDocument(doc, "https://sitename.com/album/12345")

	if err != nil {
		t.Fatalf("parseDocument() error = %v", err)
	}

	// Check title
	if result.Torrent.Title != "Goldberg Variations" {
		t.Errorf("Title = %q, want %q", result.Torrent.Title, "Goldberg Variations")
	}

	// Check year
	if result.Torrent.OriginalYear != 1981 {
		t.Errorf("Year = %d, want %d", result.Torrent.OriginalYear, 1981)
	}

	// Check edition
	if result.Torrent.Edition == nil {
		t.Fatal("Edition is nil")
	}
	if result.Torrent.Edition.Label != "Sony Classical" {
		t.Errorf("Label = %q, want %q", result.Torrent.Edition.Label, "Sony Classical")
	}

	// Check tracks
	if len(result.Torrent.Files) != 1 {
		t.Fatalf("Track count = %d, want 1", len(result.Torrent.Files))
	}

	track, ok := result.Torrent.Files[0].(*domain.Track)
	if !ok {
		t.Fatal("First file is not a track")
	}

	if track.Title != "Aria" {
		t.Errorf("Track title = %q, want %q", track.Title, "Aria")
	}
}
```

### Step 2: Run Tests

```bash
# Run tests
go test ./internal/scraping -v -run TestSiteName

# Run with coverage
go test ./internal/scraping -cover -run TestSiteName

# Run all scraping tests
go test ./internal/scraping -v
```

### Step 3: Integration Test (Optional)

Create network-dependent test (skippable in CI):

```go
func TestSiteNameExtractor_Extract_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	e := NewSiteNameExtractor()
	result, err := e.Extract("https://sitename.com/album/12345")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result.Torrent.Title == "" {
		t.Error("Title is empty")
	}
	if len(result.Torrent.Files) == 0 {
		t.Error("No tracks extracted")
	}

	t.Logf("Extracted: %s (%d tracks)", result.Torrent.Title, len(result.Torrent.Files))
}
```

---

## Integration

### Step 1: Register Extractor

In `cmd/extract/main.go`:

```go
func main() {
	// ... existing code ...

	// Create extractor registry
	registry := scraping.NewRegistry()
	
	// Register all extractors
	registry.Register(scraping.NewDiscogsExtractor())
	registry.Register(scraping.NewHarmoniaMundiExtractor())
	registry.Register(scraping.NewSiteNameExtractor()) // Add your extractor

	// ... rest of main ...
}
```

### Step 2: Test Integration

```bash
# Test with your new extractor
./extract --url "https://sitename.com/album/12345" --output test.json

# Verify output
cat test.json

# Validate
./validate --metadata test.json
```

---

## Best Practices

### 1. Error Handling

```go
// ✅ Good - specific errors
if title == "" {
	result.AddError("title", "Album title not found on page")
}

// ❌ Bad - vague errors
if title == "" {
	return nil, fmt.Errorf("error")
}
```

### 2. Logging Progress

```go
// Add notes for debugging
result.AddNote("Found 32 tracks across 2 discs")
result.AddNote("Multi-movement work detected")
```

### 3. Handle Missing Data Gracefully

```go
// Required fields - add error
if title == "" {
	result.AddError("title", "Required field missing")
}

// Optional fields - add warning
if catalogNumber == "" {
	result.AddWarning("catalog", "Catalog number not found")
}
```

### 4. Sanitize All Input

```go
// Always trim whitespace
title = strings.TrimSpace(title)

// Remove invalid characters
title = sanitizeFilename(title)

// Decode HTML entities
title = html.UnescapeString(title)
```

### 5. Document CSS Selectors

```go
// Good - documented selectors
func (e *SiteNameExtractor) extractTitle(doc *goquery.Document) string {
	// CSS selector: .album-header > .title
	// Example HTML: <div class="album-header"><h1 class="title">Album Name</h1></div>
	return doc.Find(".album-header > .title").First().Text()
}
```

### 6. Respect Robots.txt

```go
// Check robots.txt before implementing
// Add user-agent if required
// Respect rate limits
// Add delays if needed
```

### 7. Cache-Friendly Design

```go
// Extractors should be stateless
// No internal caching (handled by caller)
// Deterministic output
```

---

## Checklist

Before submitting your extractor:

- [ ] `CanHandle()` correctly identifies URLs
- [ ] Extracts all required fields (title, year, tracks)
- [ ] Extracts recommended fields when available
- [ ] Handles missing fields gracefully
- [ ] Handles multi-disc albums
- [ ] Handles multiple composers/performers
- [ ] Parses artist roles correctly
- [ ] Unit tests with mock HTML
- [ ] Integration test with live URL (optional, skippable)
- [ ] Documentation in code comments
- [ ] Example URLs in comments
- [ ] Follows project coding standards

---

## Resources

- **goquery Documentation:** https://github.com/PuerkitoBio/goquery
- **CSS Selectors Reference:** https://www.w3schools.com/cssref/css_selectors.asp
- **Example Extractors:** `internal/scraping/discogs_extractor.go`
- **Metadata Sources:** [Metadata Sources Reference](../development/metadata-sources.md)
- **Testing Guide:** [Testing Guide](../development/testing-guide.md)

---

## Need Help?

- **Questions:** GitHub Discussions
- **Bugs:** GitHub Issues
- **Code Review:** Submit PR and request review

---

**Last Updated:** 2025-01-XX  
**Version:** 1.0  
**Maintainer:** classical-tagger project
