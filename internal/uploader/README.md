# Classical Music Torrent Uploader

Upload properly tagged classical music torrents to Redacted, typically to trump existing torrents with incorrect metadata.

## Overview

The uploader is the final step in the classical music torrent workflow. After extracting metadata, validating compliance, and fixing tags, this tool handles the upload process while preserving existing site metadata.

## Features

- **Metadata Preservation**: Fetches and preserves existing torrent and group metadata from Redacted
- **Artist Validation**: Validates that local artists are a superset of Redacted artists with role compatibility checking
- **Domain Model**: Uses `domain.Artist` with type-safe role enums throughout the codebase
- **Smart Caching**: 24-hour cache of API responses to minimize repeated calls
- **Rate Limiting**: Built-in rate limiter respecting Redacted's API limits (10 requests/10 seconds)
- **Dry Run Mode**: Preview what would be uploaded without making changes
- **Trump Support**: Specifically designed for trumping torrents with metadata issues

## Installation

### Prerequisites

```bash
# Required: mktorrent for creating torrent files
sudo apt-get install mktorrent  # Debian/Ubuntu
brew install mktorrent           # macOS

# Required: Go 1.25 or later
go version  # Should show go1.25 or higher
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/cehbz/classical-tagger
cd classical-tagger

# Build the uploader
go build -o upload cmd/upload/main.go

# Install to PATH (optional)
sudo cp upload /usr/local/bin/
```

## Configuration

### API Key Setup

Obtain an API key from Redacted:
1. Go to your Redacted user settings
2. Navigate to "Access Settings"
3. Create a new API key with upload permissions
4. Set the scope appropriately (typically "Torrents" and "Upload")

Configure the API key:
```bash
# Option 1: Environment variable (recommended)
export REDACTED_API_KEY="your-api-key-here"

# Option 2: Pass via command line
upload --api-key "your-api-key-here" ...
```

### Cache Directory

The uploader uses XDG Base Directory specification:
```bash
# Default cache location
~/.cache/redacted-uploader/

# Override with environment variable
export XDG_CACHE_HOME=/path/to/cache
```

## Usage

### Basic Upload (Trump)

```bash
# Trump torrent 123456 with properly tagged files
upload --dir ./tagged_album --torrent 123456
```

### Dry Run

Always recommended before actual upload:
```bash
upload --dir ./tagged_album --torrent 123456 --dry-run --verbose
```

### Custom Trump Reason

```bash
upload --dir ./tagged_album --torrent 123456 \
  --reason "Fixed composer names, work groupings, and performer credits"
```

### Clear Cache

Force fresh metadata fetch:
```bash
upload --dir ./tagged_album --torrent 123456 --clear-cache
```

## Complete Workflow

### 1. Extract Metadata
```bash
extract --url "https://www.discogs.com/release/11245120" \
  --output ./album_metadata.json
```

### 2. Validate Current Files
```bash
validate --dir ./downloaded_torrent
# Review validation errors
```

### 3. Fix Tags and Filenames
```bash
tag --dir ./downloaded_torrent \
  --reference ./album_metadata.json \
  --output ./tagged_album
```

### 4. Validate Fixed Files
```bash
validate --dir ./tagged_album
# Ensure all issues are resolved
```

### 5. Upload (Trump)
```bash
# Dry run first
upload --dir ./tagged_album --torrent 123456 --dry-run --verbose

# If everything looks correct, do actual upload
upload --dir ./tagged_album --torrent 123456
```

## How It Works

### Metadata Flow

1. **Fetch from Redacted**
   - Torrent metadata (format, encoding, description)
   - Group metadata (detailed artist credits, tags)
   
2. **Load Local Metadata**
   - Tagged FLAC files (from `tag` command)
   - Extracted metadata JSON (from `extract` command)
   - Discogs data (if available)

3. **Merge with Precedence**
   - Artists from local files (superset of Redacted)
   - Site metadata (tags, format info) from Redacted
   - Audio metadata from validated/tagged files
   - Fill gaps with Discogs/extracted data

4. **Validate Consistency**
   - Local artists must be a superset of Redacted artists
   - Role compatibility checking (allows some flexibility)
   - Required fields presence
   - No conflicting metadata

5. **Upload**
   - Create .torrent file
   - Convert domain artists to API format with importance values
   - Submit with merged metadata
   - Preserve original description + trump notes

### Artist Role Mapping

| Domain Role | Redacted Importance | Redacted Role String | Notes |
|-------------|---------------------|----------------------|-------|
| Composer | 4 | composer | Work composers |
| Conductor | 5 | conductor | Conductors |
| Ensemble/Soloist/Performer | 1 | artists | Main performers |
| Guest | 2 | with | Guest/featured artists |
| Producer | 7 | producer | Producers |

The uploader uses `domain.Artist` internally, which provides type-safe role enums. When uploading, artists are converted to Redacted's format:
- All artists go in `artists[]` array
- Each artist has a corresponding `importance[]` value (1-8)
- Importance determines how Redacted categorizes the artist

### Validation Rules

The uploader enforces validation:

- **Artist Superset**: Local artists must contain all Redacted artists (can have more)
- **Role Compatibility**: Artist roles must be compatible (allows some flexibility, e.g., Redacted "artists" can match local "ensemble", "soloist", or "performer")
- **Required Fields**: Title, year, format, encoding, media, tags, at least one artist
- **Extra Artists Allowed**: Local tags can have additional artists not in Redacted (superset validation)

Validation failures block upload unless `--dry-run` is used.

## Caching

### Cache Files

```
~/.cache/redacted-uploader/
├── torrent_123456.json         # Torrent metadata
├── group_98765.json            # Group metadata
└── torrent_123456.torrent     # Generated torrent file
```

### Cache Behavior

- **TTL**: 24 hours from fetch time
- **Automatic**: Used when available and fresh
- **Manual Clear**: `--clear-cache` flag
- **Per-Item**: Each torrent/group cached separately

## Error Handling

### Common Errors

**Rate Limited**
```
Error: rate limited, retry after 5 seconds
```
Wait and retry, or check if you're running multiple instances.

**Artist Mismatch**
```
Validation error: artist "Name" with role "composer" not found in local tags (found with incompatible role)
```
Fix tags with `tag` command or investigate the discrepancy. Ensure all Redacted artists are present in local files with compatible roles.

**Missing Required Fields**
```
Error: missing required fields: tags, year
```
Ensure metadata extraction completed successfully.

**API Authentication**
```
Error: API error 401: Invalid API key
```
Check your API key and permissions.

### Troubleshooting

1. **Enable Verbose Mode**: Use `--verbose` to see detailed progress
2. **Try Dry Run**: Use `--dry-run` to test without uploading
3. **Check Cache**: Use `--clear-cache` if metadata seems stale
4. **Validate First**: Always run `validate` before uploading
5. **Review Metadata**: Check extracted JSON files are complete

## Safety Features

- **No Destructive Changes**: Never modifies source files
- **Dry Run Mode**: Test everything before upload
- **Validation Blocks**: Won't upload with validation errors
- **Cache Preservation**: Reuses API responses to minimize load
- **Rate Limiting**: Automatic compliance with API limits

## Best Practices

1. **Always Validate First**: Run `validate` on tagged files
2. **Use Dry Run**: Test with `--dry-run` before real upload
3. **Document Trump Reason**: Be specific about what was fixed
4. **Preserve Metadata**: Don't remove existing correct information
5. **Cache Wisely**: Clear cache only when metadata has changed

## Integration with Other Tools

The uploader is designed to work with:
- `extract`: Provides reference metadata
- `validate`: Ensures compliance before upload
- `tag`: Fixes issues found by validator

## Limitations

- **Trump Only**: Currently designed for trumping, not new uploads
- **Classical Focus**: Optimized for classical music metadata
- **Single Album**: Processes one album at a time
- **Requires mktorrent**: External dependency for torrent creation

## Future Enhancements

- [ ] Support for new uploads (not just trumps)
- [ ] Batch upload capability
- [ ] Integration with MusicBrainz
- [ ] Automatic validation before upload
- [ ] Upload progress indication
- [ ] Resume interrupted uploads

## Contributing

See main project CONTRIBUTING.md for guidelines.

## License

MIT License - See project LICENSE file