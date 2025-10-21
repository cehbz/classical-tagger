# Classical Music Metadata Sources

Reference guide for implementing web scrapers based on the Classical Music Guide appendix.

## Official Recommended Sources

These are the sources listed in the "Guide for Classical Music uploads" (Appendix: Classical music databases):

### 1. Classical Archives ⭐
**URL:** http://www.classicalarchives.com/  
**Priority:** HIGH (mentioned first in guide)

**Why Important:**
- Mentioned first in the guide as an example
- Most comprehensive classical music database
- Contains detailed work catalogs
- Extensive composer information

**Implementation Plan:**
```go
// internal/scraping/classicalarchives.go
type ClassicalArchivesExtractor struct {
    client *http.Client
}

func (e *ClassicalArchivesExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "classicalarchives.com")
}
```

**Estimated Effort:** 2-3 hours

---

### 2. Naxos ⭐
**URL:** http://www.naxos.com/  
**Priority:** HIGH

**Why Important:**
- Major classical music label and database
- Well-structured catalog
- Accurate metadata
- International coverage

**Implementation Plan:**
```go
// internal/scraping/naxos.go
type NaxosExtractor struct {
    client *http.Client
}

func (e *NaxosExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "naxos.com")
}
```

**Estimated Effort:** 2-3 hours

---

### 3. ArkivMusic ⭐
**URL:** http://arkivmusic.com/  
**Priority:** MEDIUM

**Why Important:**
- Comprehensive classical music retailer database
- Detailed performer information
- Edition-specific metadata
- Catalog numbers

**Implementation Plan:**
```go
// internal/scraping/arkivmusic.go
type ArkivMusicExtractor struct {
    client *http.Client
}

func (e *ArkivMusicExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "arkivmusic.com")
}
```

**Estimated Effort:** 2-3 hours

---

### 4. Presto Classical ⭐
**URL:** http://www.prestoclassical.co.uk/  
**Priority:** MEDIUM

**Why Important:**
- UK-based classical music retailer
- Good for recent releases
- Detailed track listings
- Review excerpts

**Implementation Plan:**
```go
// internal/scraping/prestoclassical.go
type PrestoClassicalExtractor struct {
    client *http.Client
}

func (e *PrestoClassicalExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "prestoclassical.co.uk")
}
```

**Estimated Effort:** 2-3 hours

---

### 5. Harmonia Mundi (Already Started) ✅
**URL:** http://www.harmoniamundi.com/  
**Priority:** MEDIUM (label-specific)

**Status:** Framework complete, needs HTML parsing

**Why Important:**
- Major classical music label
- High-quality recordings
- Detailed booklet scans (sometimes)
- Already has skeleton implementation

**Current Files:**
- `internal/scraping/harmoniamund.go` ✅ (framework)
- `internal/scraping/harmoniamund_test.go` ✅

**To Complete:**
- Implement `parseHTML()` method
- Add CSS selectors
- Test with live URLs

**Estimated Effort:** 1-2 hours

---

## Implementation Order

### Phase 1: Core Sources (Essential)
1. **Classical Archives** - Most comprehensive
2. **Harmonia Mundi** - Already started
3. **Naxos** - Major label

**Target:** 80% coverage of classical music uploads

### Phase 2: Additional Sources (Complete Coverage)
4. **ArkivMusic** - Retailer perspective
5. **Presto Classical** - UK releases

**Target:** 95% coverage

### Phase 3: Optional Sources (Future)
- AllMusic Classical section
- Discogs (classical-specific)
- MusicBrainz
- Label-specific sites (DG, Decca, Sony Classical)

---

## Common Fields to Extract

All scrapers should extract:

### Required Fields
- Album title
- Original year
- Composer(s)
- Track titles
- Track numbers
- Disc numbers (if multi-disc)

### Recommended Fields
- Label name
- Catalog number
- Edition year (if different from original)
- Performers (soloists, ensembles, conductors)
- Artist roles
- Opus/catalog numbers

### Optional Fields
- Cover art URL
- Recording date
- Recording location
- Producer
- Engineer
- Notes/description

---

## Technical Considerations

### HTML Parsing
All sites will use `goquery`:
```bash
go get github.com/PuerkitoBio/goquery
```

### Rate Limiting
Respect each site's robots.txt and implement delays:
```go
time.Sleep(1 * time.Second) // Between requests
```

### Error Handling
- Network errors → retry with backoff
- 404 errors → clear error message
- Parsing errors → partial data extraction

### Testing
- Mock HTML for unit tests
- Live URL tests (skipped in CI)
- Handle site structure changes gracefully

---

## CSS Selector Discovery Process

For each new site:

1. **Study the page structure**
   ```bash
   curl -s "https://example.com/album/123" > test.html
   # Open in browser DevTools
   ```

2. **Identify key elements**
   - Album title selector
   - Track list container
   - Composer fields
   - Artist/performer fields
   - Edition information

3. **Document selectors**
   ```go
   // Example for site X:
   // .product-title          → Album title
   // .release-year           → Year
   // .track-list .track      → Each track
   // .track-title            → Track title
   // .composer-name          → Composer
   ```

4. **Test extraction**
   ```go
   doc.Find(".album-title").Text()
   doc.Find(".track-list .track").Each(...)
   ```

---

## Integration with Extract CLI

After implementing each scraper, register it:

```go
// cmd/extract/main.go
registry := scraping.DefaultRegistry()

// Add all extractors
registry.Register(scraping.NewClassicalArchivesExtractor())
registry.Register(scraping.NewNaxosExtractor())
registry.Register(scraping.NewArkivMusicExtractor())
registry.Register(scraping.NewPrestoClassicalExtractor())
registry.Register(scraping.NewHarmoniaMundiExtractor())
```

Then users can extract from any supported site:
```bash
extract -url https://www.classicalarchives.com/... -output album.json
extract -url https://www.naxos.com/... -output album.json
extract -url https://arkivmusic.com/... -output album.json
```

---

## Quality Checklist

For each scraper implementation:

- [ ] CanHandle() correctly identifies URLs
- [ ] Extracts all required fields
- [ ] Extracts recommended fields when available
- [ ] Handles missing fields gracefully
- [ ] Handles multi-disc albums
- [ ] Handles multiple composers
- [ ] Handles multiple performers
- [ ] Parses artist roles correctly
- [ ] Unit tests with mock HTML
- [ ] Integration test with live URL (skippable)
- [ ] Documentation in code comments
- [ ] Example URLs in comments

---

## Expected Output

Each scraper should produce consistent `AlbumData`:

```go
AlbumData{
    Title:        "Goldberg Variations",
    OriginalYear: 1981,
    Edition: &EditionData{
        Label:         "Sony Classical",
        CatalogNumber: "SMK89245",
        EditionYear:   1981,
    },
    Tracks: []TrackData{
        {
            Disc:     1,
            Track:    1,
            Title:    "Aria",
            Composer: "Johann Sebastian Bach",
            Artists: []ArtistData{
                {Name: "Glenn Gould", Role: "soloist"},
            },
        },
        // ... more tracks
    },
}
```

---

## Timeline

| Scraper | Effort | Dependencies |
|---------|--------|--------------|
| Harmonia Mundi | 1-2 hours | goquery |
| Classical Archives | 2-3 hours | goquery |
| Naxos | 2-3 hours | goquery |
| ArkivMusic | 2-3 hours | goquery |
| Presto Classical | 2-3 hours | goquery |
| **Total** | **10-15 hours** | - |

---

## Resources

### Documentation
- goquery: https://github.com/PuerkitoBio/goquery
- CSS Selectors: https://www.w3schools.com/cssref/css_selectors.asp
- robots.txt: Check each site's `/robots.txt`

### Testing
- Mock HTML in test files
- Use `testdata/` for fixtures
- Skip network tests in CI: `t.Skip("requires network")`

### Examples
See `internal/scraping/harmoniamund.go` for:
- Extractor structure
- Error handling patterns
- Documentation style
- Implementation notes

---

**Status:** Ready for implementation  
**Next Step:** Add goquery and complete Harmonia Mundi scraper  
**Goal:** Support all 5 official metadata sources from Classical Music Guide
