# TODO - Classical Music Tagger

## High Priority (Blocks Production Use)

### 1. Add FLAC Tag Writing Library ⚠️
**Status:** Interface complete, needs implementation

**Options:**
- [ ] Option A: Shell out to `metaflac` CLI (simplest, 10 min)
- [ ] Option B: Add `github.com/go-flac/flacvorbis` library
- [ ] Option C: Add `github.com/mewkiz/flac` library
- [ ] Option D: Implement vorbis comment writing from scratch

**Files to modify:**
- `internal/tagging/writer.go` - Update `WriteTrack()` method
- Add integration tests with real FLAC files

**Estimated effort:** 10 minutes (Option A) to 4 hours (Option D)

### 2. Add HTML Parsing Library for Web Scraping ⚠️
**Status:** Framework complete, needs HTML parsing

**Recommended:**
- [ ] Add `github.com/PuerkitoBio/goquery` (most popular)
- [ ] Or use `golang.org/x/net/html` (stdlib)

**Files to modify:**
- `internal/scraping/harmoniamund.go` - Implement `parseHTML()`

**Estimated effort:** 1-2 hours per site

---

## Medium Priority (Additional Metadata Sources)

### 3. Implement Additional Web Scrapers 📚

Based on the Classical Music Guide appendix, these are the recommended metadata sources:

#### ArkivMusic ⭐
- [ ] Create `internal/scraping/arkivmusic.go`
- [ ] Create `internal/scraping/arkivmusic_test.go`
- [ ] URL: http://arkivmusic.com/
- [ ] Study site structure and identify CSS selectors
- [ ] Register in `DefaultRegistry()`

**Priority:** Medium
**Estimated effort:** 2-3 hours

#### Classical Archives ⭐
- [ ] Create `internal/scraping/classicalarchives.go`
- [ ] Create `internal/scraping/classicalarchives_test.go`
- [ ] URL: http://www.classicalarchives.com/
- [ ] Study site structure and identify CSS selectors
- [ ] Register in `DefaultRegistry()`

**Priority:** Medium (mentioned first in guide)
**Estimated effort:** 2-3 hours

#### Naxos ⭐
- [ ] Create `internal/scraping/naxos.go`
- [ ] Create `internal/scraping/naxos_test.go`
- [ ] URL: http://www.naxos.com/
- [ ] Study site structure and identify CSS selectors
- [ ] Register in `DefaultRegistry()`

**Priority:** Medium
**Estimated effort:** 2-3 hours

#### Presto Classical ⭐
- [ ] Create `internal/scraping/prestoclassical.go`
- [ ] Create `internal/scraping/prestoclassical_test.go`
- [ ] URL: http://www.prestoclassical.co.uk/
- [ ] Study site structure and identify CSS selectors
- [ ] Register in `DefaultRegistry()`

**Priority:** Medium
**Estimated effort:** 2-3 hours

#### Complete Harmonia Mundi (Started) ✅
- [ ] Implement HTML parsing in `harmoniamund.go`
- [ ] Add CSS selectors for all fields
- [ ] Test with live URLs
- [ ] Handle edge cases (multi-disc, missing fields)

**Priority:** High (already started)
**Estimated effort:** 1-2 hours

### Scraper Implementation Template

For each new scraper, follow this pattern:

```go
// internal/scraping/sitename.go
package scraping

type SiteNameExtractor struct {
    client *http.Client
}

func NewSiteNameExtractor() *SiteNameExtractor {
    return &SiteNameExtractor{
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

func (e *SiteNameExtractor) Name() string {
    return "Site Name"
}

func (e *SiteNameExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "sitename.com")
}

func (e *SiteNameExtractor) Extract(url string) (*AlbumData, error) {
    // 1. Fetch HTML
    // 2. Parse with goquery
    // 3. Extract fields
    // 4. Return AlbumData
}
```

---

## Important (Quality Improvements)

### 4. Enhanced Artist Parsing
- [ ] Parse "Soloist, Ensemble, Conductor" format properly
- [ ] Handle complex artist strings
- [ ] Auto-detect roles when not specified
- [ ] Support multiple soloists
- [ ] Test with various artist formats

**Files to modify:**
- `internal/tagging/reader.go` - Enhance `ToTrack()` method
- Add comprehensive test cases

**Estimated effort:** 2-3 hours

### 5. Arranger Detection
- [ ] Parse "(arr. by X)" from track titles
- [ ] Extract arranger name
- [ ] Add to track artists with RoleArranger
- [ ] Test with various formats: "arr. by", "arranged by", "arr."

**Files to modify:**
- `internal/tagging/reader.go` - Add arranger parsing
- `internal/validation/validator.go` - Validate arranger format

**Estimated effort:** 1-2 hours

### 6. Title Case Validation
- [ ] Implement title case checking algorithm
- [ ] Handle exceptions (and, of, the, in, etc.)
- [ ] Allow all-caps for acronyms (CD, DVD, BBC)
- [ ] Add as WARNING level validation

**Files to modify:**
- `internal/validation/validator.go` - Add title case check
- Add test cases for various capitalization styles

**Estimated effort:** 2-3 hours

### 7. Movement Format Validation
- [ ] Validate opus/catalog numbers format
- [ ] Check movement numbers (I., II., III. or 1., 2., 3.)
- [ ] Ensure consistency across multi-movement works
- [ ] Add as INFO level suggestions

**Files to modify:**
- `internal/validation/validator.go` - Add movement validation
- Add comprehensive test cases

**Estimated effort:** 2-3 hours

### 8. Test Fixtures
- [ ] Create test FLAC files with various tag combinations
- [ ] Add to `testdata/` directory
- [ ] Update integration tests to use fixtures
- [ ] Document fixture creation process

**Files to create:**
- `testdata/` directory with sample FLAC files
- `testdata/README.md` - Documentation

**Estimated effort:** 1-2 hours

---

## Nice to Have (Future Enhancements)

### 9. Configuration System
- [ ] User preferences file (~/.classical-tagger.yaml)
- [ ] Configurable validation levels
- [ ] Custom rules
- [ ] Preferred metadata sources
- [ ] Output formatting preferences

**Estimated effort:** 3-4 hours

### 10. Batch Processing
- [ ] Process multiple albums in one run
- [ ] Parallel extraction/validation
- [ ] Progress bars
- [ ] Summary reports

**Estimated effort:** 4-6 hours

### 11. Cover Art Support
- [ ] Download cover art from websites
- [ ] Embed in FLAC files
- [ ] Validate image dimensions
- [ ] Extract from existing files

**Estimated effort:** 3-4 hours

### 12. Additional Audio Formats
- [ ] MP3 support
- [ ] M4A support
- [ ] DSD/DSF support
- [ ] Format detection

**Estimated effort:** 6-8 hours

### 13. Undo/Rollback System
- [ ] Track all changes
- [ ] Rollback command
- [ ] Change history
- [ ] Selective undo

**Estimated effort:** 4-6 hours

### 14. Web UI (Optional)
- [ ] Simple web interface for validation
- [ ] Drag-and-drop album folders
- [ ] Visual diff for changes
- [ ] Batch operations

**Estimated effort:** 20+ hours

---

## Documentation

### 15. User Guide
- [ ] Installation instructions
- [ ] Complete workflow examples
- [ ] Troubleshooting guide
- [ ] FAQ section
- [ ] Video tutorials (optional)

**Estimated effort:** 4-6 hours

### 16. API Documentation
- [ ] GoDoc comments for all public APIs
- [ ] Architecture diagrams
- [ ] Integration examples
- [ ] Contributing guide

**Estimated effort:** 2-3 hours

---

## Metadata Source Priority

Based on the Classical Music Guide, implement scrapers in this order:

1. **Classical Archives** (mentioned first, most comprehensive)
2. **Naxos** (large catalog, well-structured)
3. **ArkivMusic** (comprehensive metadata)
4. **Presto Classical** (good for recent releases)
5. **Harmonia Mundi** (already started, label-specific)

## Estimated Total Effort

| Priority | Tasks | Est. Hours |
|----------|-------|------------|
| High | 2 items | 2-5 hours |
| Medium | 5 scrapers | 10-15 hours |
| Important | 5 items | 12-18 hours |
| Nice to Have | 9 items | 50+ hours |
| **Total** | **21 items** | **74-88+ hours** |

## Quick Wins (Do First)

1. ✅ Add metaflac integration (10 min)
2. ✅ Complete Harmonia Mundi HTML parsing (1-2 hours)
3. ✅ Add goquery library (5 min)
4. ✅ Implement Classical Archives scraper (2-3 hours)
5. ✅ Enhanced artist parsing (2-3 hours)

**Total Quick Wins:** ~6-9 hours to get to fully functional state

## Current Status Summary

| Component | Status | Next Step |
|-----------|--------|-----------|
| Domain Model | ✅ Complete | - |
| Validation | ✅ Complete | Add title case validation |
| Tag Reading | ✅ Complete | Enhance artist parsing |
| Tag Writing | ⚠️ 95% | Add FLAC library |
| Filesystem | ✅ Complete | - |
| Storage | ✅ Complete | - |
| Scraping Framework | ✅ Complete | Implement HTML parsing |
| Harmonia Mundi | ⚠️ 80% | Complete HTML parsing |
| Other Scrapers | ⏸️ Not Started | Implement 4 additional sources |
| Validate CLI | ✅ Complete | - |
| Tag CLI | ⚠️ 95% | Add FLAC library |
| Extract CLI | ⚠️ 90% | Complete HTML parsing |

## Notes

- All framework code is complete and tested
- Two libraries needed for full functionality (FLAC writing, HTML parsing)
- Five additional scrapers requested based on Classical Music Guide
- Project is ~80% complete, ~20 hours from production-ready
- Additional scrapers add ~10-15 hours

## Getting Started

To continue development:

```bash
# 1. Add libraries
go get github.com/PuerkitoBio/goquery

# 2. Implement metaflac integration (10 min)
# Edit internal/tagging/writer.go

# 3. Complete Harmonia Mundi scraper (1-2 hours)  
# Edit internal/scraping/harmoniamund.go

# 4. Add first new scraper - Classical Archives (2-3 hours)
# Create internal/scraping/classicalarchives.go

# 5. Test everything
go test ./...
```

## Blockers

- None currently
- All dependencies are available
- All tools are working
- Framework is complete

---

**Last Updated:** October 21, 2025
**Status:** Ready for implementation
**Next Milestone:** Complete 5 metadata source scrapers
