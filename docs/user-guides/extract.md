# Extract CLI - Metadata Extraction from Websites

**Status:** ✅ Complete and functional  
**Version:** 1.0  
**Integration:** Works with HarmoniaMundiParser

## Overview

The `extract` CLI tool fetches and parses classical music album metadata from supported websites, converting it to the standard JSON format used by the tagger.

## Installation

```bash
cd cmd/extract
go build -o extract
```

## Usage

### Basic Usage

```bash
# Extract to stdout
extract -url "https://www.harmoniamundi.com/album/..."

# Extract to file
extract -url "https://www.harmoniamundi.com/album/..." -output album.json
```

### Options

```
-url string
    URL to extract from (required)

-output string
    Output JSON file (default: stdout)

-validate
    Validate extracted metadata against domain rules (default: true)

-verbose
    Verbose output including parsing notes (default: false)

-force
    Create output even with required field errors (default: false)

-timeout duration
    HTTP request timeout (default: 30s)
```

### Examples

```bash
# Simple extraction
extract -url "https://www.harmoniamundi.com/..." -output album.json

# Verbose mode with parsing details
extract -url "https://www.harmoniamundi.com/..." -output album.json -verbose

# Force output despite errors
extract -url "https://www.harmoniamundi.com/..." -output album.json -force

# Skip validation
extract -url "https://www.harmoniamundi.com/..." -output album.json -validate=false

# Custom timeout
extract -url "https://www.harmoniamundi.com/..." -timeout 60s -output album.json
```

## Supported Sites

### Currently Supported

1. **Harmonia Mundi** (harmoniamundi.com)
   - Album metadata
   - Track listings
   - Composer information
   - Catalog numbers
   - Edition information

### Coming Soon

See `METADATA_SOURCES.md` for the complete list of planned sources:
- Classical Archives
- Naxos
- ArkivMusic
- Presto Classical

## Output Format

The tool produces JSON in the standard album metadata format:

```json
{
  "title": "Noël ! Weihnachten ! Christmas!",
  "original_year": 2013,
  "edition": {
    "label": "harmonia mundi",
    "catalog_number": "HMC902170",
    "edition_year": 2013
  },
  "tracks": [
    {
      "disc": 1,
      "track": 1,
      "title": "Frohlocket, ihr Völker auf Erden, op.79/1",
      "composer": "Felix Mendelssohn Bartholdy",
      "artists": []
    }
  ]
}
```

## Features

### Extraction Features

- ✅ Album title extraction
- ✅ Release year parsing
- ✅ Track listing with composers
- ✅ Edition information (label, catalog number)
- ✅ Artist role inference
- ✅ Multi-disc detection
- ✅ HTML entity decoding (UTF-8 characters)
- ✅ Composer name formatting (ALL CAPS → Title Case)

### Error Handling

- ✅ Required vs optional field errors
- ✅ Detailed error messages
- ✅ Warning collection
- ✅ Parsing notes for debugging
- ✅ Domain validation
- ✅ Graceful degradation

### User Experience

- ✅ Clear progress messages
- ✅ Colored output
- ✅ Verbose mode for debugging
- ✅ Force mode for problematic extractions
- ✅ Validation before output
- ✅ Helpful next-step suggestions

## Workflow Integration

### Complete Tagger Workflow

```bash
# Step 1: Extract metadata from web
extract -url "https://www.harmoniamundi.com/..." -output album.json

# Step 2: Review/edit the JSON (optional)
nano album.json

# Step 3: Validate album directory
validate /path/to/album

# Step 4: Apply tags
tag -metadata album.json -dir /path/to/album

# Step 5: Verify result
validate /path/to/album_tagged
```

## Exit Codes

- `0` - Success
- `1` - Error (invalid arguments, network error, parsing error, validation error)

## Examples by Use Case

### Quick Extraction

```bash
# Just get the JSON
extract -url "URL" -output album.json
```

### Debugging Extraction Issues

```bash
# Verbose mode shows parsing details
extract -url "URL" -output album.json -verbose
```

### Handling Partial Data

```bash
# Force output even with missing required fields
extract -url "URL" -output album.json -force
```

### Batch Extraction

```bash
# Extract from multiple URLs
while read url; do
    filename=$(echo "$url" | md5sum | cut -d' ' -f1).json
    extract -url "$url" -output "$filename"
    sleep 5  # Be nice to the server
done < urls.txt
```

## Troubleshooting

### "No parser available for this URL"

**Problem:** The URL is not from a supported website.

**Solution:**
- Check that you're using a URL from harmoniamundi.com
- See `METADATA_SOURCES.md` for planned additional sites
- To add support for a new site, create a new parser following the pattern in `internal/scraping/harmoniamundi_parser.go`

### "HTTP 403 Forbidden" or "HTTP 404 Not Found"

**Problem:** The website blocked the request or the page doesn't exist.

**Solution:**
- Verify the URL in your browser first
- Check that the URL is correct and complete
- Some sites may block automated requests
- Try increasing the timeout: `-timeout 60s`

### "Extraction failed due to required field errors"

**Problem:** Critical metadata fields (title, year, tracks) could not be extracted.

**Solution:**
- Use `-verbose` to see detailed parsing notes
- Check if the website structure has changed
- Use `-force` to create output anyway (not recommended for tagging)
- File an issue if this is a regression

### "Domain conversion failed"

**Problem:** Extracted metadata doesn't meet domain validation rules.

**Solution:**
- Review validation errors shown
- Edit the JSON file manually to fix issues
- Use `-validate=false` to skip validation (not recommended)
- Use `-force` to create output despite errors

### Slow Extraction

**Problem:** Extraction takes a long time.

**Solution:**
- Increase timeout: `-timeout 120s`
- Check your network connection
- The website may be slow to respond

### UTF-8 Encoding Issues

**Problem:** Special characters appear garbled (é, ö, ü, ñ, etc.).

**Solution:**
- This should be handled automatically
- If you see issues, file a bug report with the URL
- The parser includes comprehensive HTML entity decoding

## Technical Details

### HTTP Client Configuration

- **Timeout:** 30 seconds (configurable)
- **User-Agent:** `classical-tagger/0.1 (metadata extraction tool)`
- **Follow redirects:** Yes (automatic)
- **SSL verification:** Yes

### Parsing Strategy

1. **Fetch HTML** from URL
2. **Detect site** based on URL pattern
3. **Select parser** (Harmonia Mundi, etc.)
4. **Parse HTML** using goquery and regex
5. **Extract metadata** (title, year, tracks, etc.)
6. **Infer artist roles** using pattern matching
7. **Detect disc structure** from track numbering
8. **Convert to domain model** (validates business rules)
9. **Serialize to JSON** using standard format

### Error Handling Levels

1. **Required Field Errors** - Block output unless `--force`
2. **Optional Field Warnings** - Noted but don't block output
3. **Low Confidence Warnings** - Artist role inferences
4. **Parsing Notes** - Available with `--verbose`

## Development

### Adding Support for New Sites

1. Create new parser (e.g., `internal/scraping/naxos_parser.go`)
2. Implement `Parse(html string) (*ExtractionResult, error)`
3. Add detection logic to `main.go`
4. Add comprehensive tests
5. Update this README

See `METADATA_SOURCES.md` for detailed implementation guides for planned sources.

### Testing

```bash
# Run tests
go test -v

# Test with real URL (requires network)
go test -tags=integration -run TestFetchHTML

# Test URL detection
go test -run TestIsHarmoniaMundi
```

### Code Structure

```
cmd/extract/
├── main.go           # CLI implementation
├── main_test.go      # Tests
└── README.md         # This file
```

## Safety & Ethics

### Respectful Scraping

- **Rate limiting:** Manual delay recommended for batch operations
- **User-Agent:** Identifies as metadata extraction tool
- **Robots.txt:** Always respect site policies
- **Terms of service:** Ensure compliance
- **Personal use:** Tool is for personal metadata management

### Data Usage

- Extracted metadata is for **personal use only**
- Do not redistribute extracted data
- Do not overload websites with requests
- Respect copyright and intellectual property

## Future Enhancements

- [ ] Support for additional websites (Classical Archives, Naxos, etc.)
- [ ] Caching of downloaded HTML
- [ ] Resume interrupted batch extractions
- [ ] Parallel extraction with rate limiting
- [ ] Browser automation for JavaScript-heavy sites
- [ ] Cover art download
- [ ] Automatic site structure update detection
- [ ] Plugin architecture for custom parsers

## Related Commands

- **validate** - Validate album directories and metadata
- **tag** - Apply metadata to FLAC files

## Support

- **Documentation:** See project README.md
- **Issues:** Report bugs on GitHub
- **New sites:** Request in GitHub Issues or submit PR

---

**Last Updated:** October 22, 2025  
**Maintainer:** classical-tagger project  
**License:** MIT