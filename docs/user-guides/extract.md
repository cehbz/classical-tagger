# Extract CLI - Metadata Extraction from FLAC Files and Discogs

**Status:** ✅ Complete and functional  
**Version:** 1.0

## Overview

The `extract` CLI tool extracts metadata from FLAC files in a directory and optionally enriches it with data from Discogs. It creates two JSON files: one with local metadata extracted from FLAC tags, and another with Discogs metadata if available.

## Installation

```bash
cd cmd/extract
go build -o extract
```

## Usage

### Basic Usage

```bash
# Extract from directory with automatic Discogs lookup
extract -dir "/path/to/album"

# Extract with specific Discogs release ID
extract -dir "/path/to/album" -release-id 11245120

# Local extraction only (skip Discogs)
extract -dir "/path/to/album" -no-api
```

### Options

```
-dir string
    Directory containing FLAC files (required)

-release-id int
    Specific Discogs release ID to use (skips search)

-output string
    Base name for output files (default: directory name)

-verbose
    Enable verbose output (default: false)

-force
    Create output even if required fields are missing (default: false)

-no-api
    Skip Discogs API lookup (default: false)
```

### Examples

```bash
# Basic extraction with automatic Discogs search
extract -dir "/music/Bach - Goldberg Variations"

# Use specific Discogs release
extract -dir "/music/Bach - Goldberg Variations" -release-id 195873

# Verbose mode to see search process
extract -dir "/music/album" -verbose

# Local extraction only
extract -dir "/music/album" -no-api
```

## Discogs Integration

### Search Behavior

The extract command searches Discogs using two strategies:

1. **Advanced Search** (first attempt): Uses separate `artist` and `release_title` parameters with format restriction (CD). This is precise but strict about spelling.

2. **Simple Search** (fallback): If advanced search finds no results, automatically tries a simple search using a combined `query` parameter with both artist and album title. This is more forgiving of spelling variations and doesn't restrict by format.

**Example:** If searching for "Weinachten" fails, the fallback search with "RIAS Kammerchor Weinachten" may find the release titled "Weihnachten".

### Role Determination

When converting Discogs releases, artist roles are determined with the following priority:

1. **Discogs main artist role**: If the artist has an explicit role in Discogs main artists list, use it.
2. **Discogs extraartists role**: If the artist appears in Discogs extraartists with a role, use that role.
3. **File metadata role**: If the artist exists in the local FLAC file metadata with a role, use that role.
4. **Error**: If no role can be determined from any source, the extraction fails with an error listing which artists have unknown roles.

This ensures that roles are always properly determined and prevents silent data quality issues.

### Error Handling

If Discogs data cannot determine roles for all artists, the extraction will fail with a clear error message:

```
Error saving Discogs data: failed to convert Discogs release: cannot determine role for album artist 'Artist Name'. Discogs has no role, extraartists has no matching entry, and file metadata has no matching entry
```

**Solutions:**
- Use `--release-id` with a different Discogs release that has better role information
- Ensure your FLAC files have proper artist role tags
- Manually edit the Discogs release on discogs.com to add role information

## Output Format

The tool creates two JSON files:

1. **`<name>.json`**: Local metadata extracted from FLAC files
2. **`<name>_discogs.json`**: Metadata from Discogs API (if available)

Both files use the standard torrent metadata format:

```json
{
  "title": "Noël! Christmas! Weihnachten!",
  "original_year": 2013,
  "edition": {
    "label": "Harmonia Mundi",
    "catalog_number": "HMC 902170",
    "year": 2013
  },
  "album_artist": [
    {"name": "RIAS-Kammerchor", "role": "ensemble"},
    {"name": "Hans-Christoph Rademann", "role": "conductor"}
  ],
  "files": [
    {
      "disc": 1,
      "track": 1,
      "title": "Frohlocket, Ihr Völker Auf Erden (op.79/1)",
      "artists": [
        {"name": "Felix Mendelssohn-Bartholdy", "role": "composer"},
        {"name": "RIAS-Kammerchor", "role": "ensemble"},
        {"name": "Hans-Christoph Rademann", "role": "conductor"}
      ]
    }
  ]
}
```

## Features

### Extraction Features

- ✅ Metadata extraction from FLAC files
- ✅ Album title, year, and track information from tags
- ✅ Artist and composer information
- ✅ Discogs API integration for metadata enrichment
- ✅ Automatic Discogs search with fallback
- ✅ Role determination from multiple sources
- ✅ Multi-disc detection
- ✅ Edition information (label, catalog number)

### Discogs Integration

- ✅ Automatic search by artist and album title
- ✅ Fallback to simple search if advanced search fails
- ✅ Support for specific release ID lookup
- ✅ Role determination with priority: Discogs > file metadata
- ✅ Validation to ensure all roles are determined
- ✅ Clear error messages when roles cannot be determined

### Error Handling

- ✅ Required vs optional field errors
- ✅ Detailed error messages for role determination failures
- ✅ Warning collection
- ✅ Verbose mode for debugging search process
- ✅ Domain validation
- ✅ Graceful degradation (can skip Discogs with `--no-api`)

### User Experience

- ✅ Clear progress messages
- ✅ Verbose mode shows search attempts
- ✅ Force mode for problematic extractions
- ✅ Helpful error messages with solutions

## Workflow Integration

### Complete Tagger Workflow

```bash
# Step 1: Extract metadata from FLAC files and Discogs
extract -dir "/path/to/album" -verbose

# This creates:
# - album.json (local metadata)
# - album_discogs.json (Discogs metadata)

# Step 2: Validate the Discogs metadata
validate album_discogs.json album.json

# Step 3: Apply tags using Discogs metadata
tag -metadata album_discogs.json -dir /path/to/album -output /path/to/tagged

# Step 4: Verify result
validate /path/to/tagged
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

### "No Discogs releases found"

**Problem:** Discogs search found no matching releases.

**Solutions:**
- Check that your FLAC files have proper artist and album tags
- Try using `-release-id` with a specific Discogs release ID if you know it
- The fallback simple search should help with spelling variations
- Use `-verbose` to see what search terms are being used
- Check Discogs manually to find the correct release ID

### "Cannot determine role for artist"

**Problem:** Discogs and file metadata don't have role information for an artist.

**Solutions:**
- Use `--release-id` with a different Discogs release that has better role information
- Ensure your FLAC files have proper artist role tags (ARTIST, COMPOSER, etc.)
- Manually edit the Discogs release on discogs.com to add role information
- The error message will list which artists need roles

### "Cannot search Discogs without artist and album information"

**Problem:** FLAC files don't have sufficient metadata to search Discogs.

**Solutions:**
- Ensure your FLAC files have ALBUM and ARTIST tags
- Use `--no-api` to skip Discogs lookup and work with local metadata only
- Use `--release-id` if you know the Discogs release ID

### "Discogs search failed"

**Problem:** Network error or API issue when searching Discogs.

**Solutions:**
- Check your internet connection
- Verify your Discogs API token is configured in `~/.config/classical-tagger/config.yaml`
- Check Discogs API status
- Use `--no-api` to skip Discogs lookup temporarily

### "Error saving Discogs data"

**Problem:** Failed to convert or save Discogs release data.

**Solutions:**
- Check the error message for specific issues (usually role determination)
- Use `-verbose` to see more details
- Try a different release ID if roles cannot be determined
- Ensure file permissions allow writing to the output directory

## Technical Details

### Discogs API Configuration

- **Rate Limiting:** 60 requests per minute (automatic)
- **Caching:** Search results and release data are cached
- **User-Agent:** `ClassicalTagger/1.0`
- **Authentication:** Requires Discogs API token in config file

### Extraction Process

1. **Extract local metadata** from FLAC files in directory
2. **Save local metadata** to `<name>.json`
3. **Search Discogs** using artist and album from local metadata
   - First tries advanced search (artist + release_title + format)
   - Falls back to simple search (query parameter) if no results
4. **Fetch release details** if single match found
5. **Determine artist roles** with priority:
   - Discogs main artist role
   - Discogs extraartists role
   - File metadata role
6. **Validate roles** - fail if any artist has unknown role
7. **Convert to domain model** and save to `<name>_discogs.json`

### Role Determination Priority

When converting Discogs releases, roles are determined in this order:

1. **Discogs main artist role**: Explicit role in `release.Artists[].role`
2. **Discogs extraartists role**: Role from `release.ExtraArtists[]` if artist name matches
3. **File metadata role**: Role from local FLAC tags if artist name matches
4. **Error**: Extraction fails if no role can be determined

This ensures data quality and prevents silent role assignment issues.

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