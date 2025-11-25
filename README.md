# Classical Music Tagger

A comprehensive toolkit for managing classical music torrents on Redacted. Validates metadata compliance, extracts information from authoritative sources, fixes tags, and handles uploads—all following strict classical music guidelines.

## What Does This Do?

This toolkit helps you:
1. **Validate** existing torrents against classical music rules
2. **Extract** metadata from authoritative classical music sources
3. **Tag** FLAC files with proper classical metadata
4. **Upload** corrected torrents to Redacted (trump bad uploads)

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/cehbz/classical-tagger
cd classical-tagger

# Build all tools
go build -o validate cmd/validate/main.go
go build -o extract cmd/extract/main.go
go build -o tag cmd/tag/main.go
go build -o upload cmd/upload/main.go

# Optional: Install to PATH
sudo cp validate extract tag upload /usr/local/bin/
```

### Configuration

Create `~/.config/classical-tagger/config.yaml`:

```yaml
# Discogs API token (get from https://www.discogs.com/settings/developers)
discogs:
  token: "your-discogs-token-here"

# Redacted API key (get from https://redacted.sh/user.php?action=edit)
redacted:
  api_key: "your-redacted-api-key-here"

# Optional: Cache TTL in hours (default: 24)
cache:
  ttl_hours: 24
```

### Your First Workflow

```bash
# 1. Extract metadata from Discogs
extract --url "https://www.discogs.com/release/12345" --output metadata.json

# 2. Validate a downloaded torrent
validate --dir ./downloaded-torrent

# 3. Fix the tags using extracted metadata
tag --metadata metadata.json --dir ./downloaded-torrent --output ./fixed-torrent

# 4. Upload to trump the bad torrent
upload --dir ./fixed-torrent --torrent 123456
```

## Command Overview

### validate
Check torrents for compliance with classical music rules.

```bash
validate --dir ./my-album
```

**Key Features:**
- Filename format validation
- Tag completeness checking
- Classical music-specific rules
- Multi-disc support
- Detailed error reports

[Full Documentation](docs/user-guides/validate-guide.md)

### extract
Fetch metadata from authoritative sources.

```bash
extract --url "https://www.discogs.com/..." --output album.json
```

**Supported Sources:**
- Discogs (implemented)
- Harmonia Mundi (in progress)
- Classical Archives (planned)
- Naxos (planned)
- Presto Classical (planned)

[Full Documentation](docs/user-guides/extract-guide.md)

### tag
Apply metadata to FLAC files with proper formatting.

```bash
tag --metadata album.json --dir ./source --output ./tagged
```

**Key Features:**
- Non-destructive (originals untouched)
- Validates before applying
- Multi-disc directory structure
- Dry-run mode
- Automatic backups

[Full Documentation](docs/user-guides/tag-guide.md)

### upload
Upload corrected torrents to Redacted.

```bash
upload --dir ./fixed-torrent --torrent 123456
```

**Key Features:**
- Preserves site metadata
- Validates artist consistency
- Smart caching (24-hour TTL)
- Rate limiting compliance
- Dry-run mode

[Full Documentation](docs/user-guides/upload-guide.md)

## Classical Music Rules

This toolkit enforces the rules from the [Redacted Classical Music Upload Guide](https://redacted.sh/wiki.php?action=article&id=197):

- ✅ Composer names in standardized format
- ✅ Work titles with opus/catalog numbers
- ✅ Multi-movement work groupings
- ✅ Performer roles (Composer, Soloist, Ensemble, Conductor)
- ✅ 180-character filename limit
- ✅ Specific filename format: `## - Track Title.flac`
- ✅ Multi-disc subdirectories

## Project Structure

```
classical-tagger/
├── cmd/                    # Command-line applications
│   ├── validate/          # Validation tool
│   ├── extract/           # Metadata extraction tool
│   ├── tag/               # Tagging tool
│   └── upload/            # Upload tool
├── internal/
│   ├── domain/            # Domain models (Album, Track, Artist)
│   ├── validation/        # Validation rules engine
│   ├── tagging/           # FLAC tag reading/writing
│   ├── scraping/          # Web metadata extraction
│   ├── config/            # Configuration management
│   └── uploader/          # Redacted upload logic
└── docs/                  # Documentation
```

## Documentation

### For Users
- **[Getting Started Guide](docs/user-guides/getting-started.md)** - Detailed setup and first workflow
- **[Validate](docs/user-guides/validate.md)** - Validation tool reference
- **[Extract](docs/user-guides/extract.md)** - Extraction tool reference
- **[Tag](docs/user-guides/tag.md)** - Tagging tool reference
- **[Upload](docs/user-guides/upload.md)** - Upload tool reference
- **[Troubleshooting](docs/user-guides/troubleshooting.md)** - Common errors and solutions

### For Developers
- **[Architecture](docs/project/ARCHITECTURE.md)** - Domain model and design principles
- **[Development Guide](docs/project/DEVELOPMENT.md)** - Coding standards and TDD workflow
- **[Metadata Sources](docs/development/metadata-sources.md)** - Reference for supported sites
- **[Adding Scrapers](docs/development/adding-scrapers.md)** - How to add new metadata sources
- **[Testing Guide](docs/development/testing-guide.md)** - Test patterns and fixtures
- **[TODO](docs/project/TODO.md)** - Future plans and priorities

## Requirements

- **Go 1.25+** - For building and running
- **FLAC files** - For tagging operations
- **mktorrent** - For creating torrent files (upload command only)
- **API Keys**:
  - Discogs personal access token (for metadata extraction)
  - Redacted API key (for upload operations)

## Design Principles

- **Domain-Driven Design (DDD)** - Clean separation of concerns
- **Test-Driven Development (TDD)** - Comprehensive test coverage
- **SOLID Principles** - Maintainable, extensible code
- **Immutability** - Domain objects are immutable where possible
- **Type Safety** - Leverage Go's type system
- **XDG Compliance** - Standard configuration paths

## Development

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/domain

# Run with coverage
go test -cover ./...

# Build all commands
go build ./cmd/...
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development guidelines.

## Contributing

We welcome contributions! Please see [DEVELOPMENT.md](DEVELOPMENT.md) for:
- Coding standards
- Testing requirements
- Pull request process
- Architecture guidelines

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Support

- **Issues**: Report bugs on [GitHub Issues](https://github.com/cehbz/classical-tagger/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/cehbz/classical-tagger/discussions)
- **Documentation**: Check the [docs/](docs/) directory

## Acknowledgments

- Based on the [Redacted Classical Music Upload Guide](https://redacted.sh/wiki.php?action=article&id=197)
- Metadata sources: Discogs, Harmonia Mundi, Classical Archives, Naxos, Presto Classical
- Community feedback from Redacted classical music enthusiasts