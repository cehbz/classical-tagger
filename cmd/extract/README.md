## extract - Web Metadata Extractor

Extracts classical music album metadata from supported websites and converts to JSON format.

## Features

- ✅ **Web scraping framework** - Extensible extractor system
- ✅ **Harmonia Mundi support** - Extract from harmoniamundi.com (framework ready)
- ✅ **Automatic validation** - Validates extracted metadata
- ✅ **JSON output** - Compatible with tag CLI
- ✅ **Error handling** - Clear error messages

## Installation

```bash
cd cmd/extract
go build -o extract
```

## Usage

### Basic Usage

```bash
# Extract to stdout
./extract -url https://www.harmoniamundi.com/en/album/123

# Save to file
./extract -url https://www.harmoniamundi.com/en/album/123 -output album.json

# Skip validation
./extract -url URL -validate=false

# Verbose output
./extract -url URL -verbose
```

### Complete Workflow

```bash
# 1. Extract metadata
./extract -url https://www.harmoniamundi.com/en/album/noel-weihnachten-christmas \
  -output metadata.json

# 2. Review and edit the JSON if needed
vi metadata.json

# 3. Validate the album directory
cd ../validate
./validate /path/to/album

# 4. Apply the metadata
cd ../tag
./tag -metadata metadata.json -dir /path/to/album
```

## Supported Sites

### ✅ Harmonia Mundi (Framework Ready)
- Domain: harmoniamundi.com
- Status: Interface implemented, HTML parsing needed
- Example: `https://www.harmoniamundi.com/en/album/...`

### 🚧 Coming Soon
- Classical Archives (classicalarchives.com)
- Naxos (naxos.com)
- Presto Classical (prestoclassical.co.uk)
- ArkivMusic (arkivmusic.com)

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | *required* | URL to extract from |
| `-output` | stdout | Output file path |
| `-validate` | `true` | Validate extracted metadata |
| `-verbose` | `false` | Show detailed output |

## Output Example

```
Finding extractor for: https://www.harmoniamundi.com/...
Using extractor: Harmonia Mundi

Extracting metadata from Harmonia Mundi...
✓ Extracted: Noël ! Weihnachten ! Christmas! (2013)
  Tracks: 24
  Label: harmonia mundi
  Catalog: HMC902170

Validating extracted metadata...
✓ Metadata is valid

Converting to JSON...
✓ Saved to: /path/to/album.json

Next steps:
  1. Review and edit: album.json
  2. Validate album: validate /path/to/album
  3. Apply tags: tag -metadata album.json -dir /path/to/album
```

## JSON Output Format

The output is compatible with the `tag` CLI:

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
      "title": "Frohlocket, ihr Völker auf Erden, Op. 79/1",
      "composer": {
        "name": "Felix Mendelssohn Bartholdy",
        "role": "composer"
      },
      "artists": [
        {
          "name": "RIAS Kammerchor Berlin",
          "role": "ensemble"
        },
        {
          "name": "Hans-Christoph Rademann",
          "role": "conductor"
        }
      ],
      "name": "01 Frohlocket, ihr Völker auf Erden, Op. 79-1.flac"
    }
  ]
}
```

## Error Handling

### Unsupported URL

```
Error: No extractor available for this URL
Supported sites:
  - harmoniamundi.com
```

### Extraction Failed

```
Error extracting metadata: HTTP 404: Not Found
```

### Validation Warnings

```
Validating extracted metadata...
⚠️  [WARNING] Track 1: Consider using catalog number format
⚠️  Extracted metadata has validation errors
You may need to manually fix the JSON before using it
```

## Important Notes

### ⚠️ Current Status

The extraction framework is **complete**, but the HTML parsing for Harmonia Mundi needs to be implemented:

✅ **What's Working:**
- Extractor interface and registry
- URL detection and routing
- Data conversion to domain model
- JSON serialization
- Validation integration
- CLI interface

❌ **What Needs Implementation:**
- Actual HTML parsing for each site
- CSS selector mapping
- Multi-disc detection from HTML
- Artist role parsing

### Implementing HTML Parsing

To complete the Harmonia Mundi extractor:

1. **Add HTML parsing library:**
   ```bash
   go get github.com/PuerkitoBio/goquery
   ```

2. **Study the website:**
   - Open browser DevTools
   - Inspect album page structure
   - Identify CSS selectors for each field

3. **Update `internal/scraping/harmoniamund.go`:**
   ```go
   func (e *HarmoniaMundiExtractor) parseHTML(html string, url string) (*AlbumData, error) {
       doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
       if err != nil {
           return nil, err
       }
       
       albumData := &AlbumData{
           Title: doc.Find(".album-title").Text(),
           // ... extract other fields
       }
       
       return albumData, nil
   }
   ```

See detailed comments in `harmoniamund.go` for implementation guidance.

## Adding New Extractors

To add support for a new website:

### 1. Create extractor file

```go
// internal/scraping/naxos.go
package scraping

type NaxosExtractor struct {
    client *http.Client
}

func NewNaxosExtractor() *NaxosExtractor {
    return &NaxosExtractor{
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

func (e *NaxosExtractor) Name() string {
    return "Naxos"
}

func (e *NaxosExtractor) CanHandle(url string) bool {
    return strings.Contains(url, "naxos.com")
}

func (e *NaxosExtractor) Extract(url string) (*AlbumData, error) {
    // Implement extraction logic
}
```

### 2. Register in CLI

```go
// cmd/extract/main.go
registry.Register(scraping.NewNaxosExtractor())
```

### 3. Add tests

```go
// internal/scraping/naxos_test.go
func TestNaxosExtractor(t *testing.T) {
    // Test implementation
}
```

## Development Workflow

### Testing with Mock Data

```bash
# Create mock HTML file
cat > test.html << 'EOF'
<html>
<h1 class="album-title">Test Album</h1>
...
</html>
EOF

# Test parser locally
go test ./internal/scraping -v -run TestParseHTML
```

### Testing with Real URLs

```bash
# Test extraction (requires network)
go test ./internal/scraping -v -run TestExtract -args -live

# Or use the CLI
./extract -url "https://www.harmoniamundi.com/..." -verbose
```

## Rate Limiting

The extractor respects websites with:
- 30 second timeout per request
- No concurrent requests by default
- User-Agent header

For batch extraction, add delays:

```bash
# Extract multiple albums with delays
for url in $(cat urls.txt); do
    ./extract -url "$url" -output "$(basename $url).json"
    sleep 5  # Be nice to the server
done
```

## Privacy & Ethics

- **Respect robots.txt** - Check before scraping
- **Rate limiting** - Don't overload servers
- **Terms of service** - Ensure compliance
- **Personal use** - This tool is for personal metadata management

## Troubleshooting

### "No extractor available"

- Check URL format
- Ensure domain is supported
- Try with `-verbose` for details

### "HTTP 403 Forbidden"

- Website may block automated requests
- Try different User-Agent
- May need browser automation (Selenium)

### "HTML parsing failed"

- Website structure may have changed
- Check CSS selectors
- Inspect page with browser DevTools

## Future Enhancements

- [ ] Complete Harmonia Mundi HTML parsing
- [ ] Add Naxos extractor
- [ ] Add Classical Archives extractor
- [ ] Batch URL processing
- [ ] Caching extracted data
- [ ] Browser automation for JavaScript sites
- [ ] Image download (cover art)
- [ ] Parallel extraction
- [ ] Resume interrupted extractions

## Integration

The extract CLI integrates seamlessly with other commands:

```bash
# Complete workflow
extract -url URL -output album.json
validate /path/to/album
tag -metadata album.json -dir /path/to/album
validate /path/to/album  # Verify
```

All three CLIs use the same JSON format for maximum compatibility.
