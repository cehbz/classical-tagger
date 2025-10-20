# Classical Music Tagger

Go applications for validating, tagging, and extracting metadata for classical music torrents.

## Commands

- `validate` - Validates tags in a downloaded torrent directory
- `tag` - Applies JSON metadata to audio files  
- `extract` - Extracts JSON metadata from web pages

## Getting Started

### Prerequisites

- Go 1.25 or later
- FLAC audio files

### Running Tests

```bash
go test ./...
```

### Running a Specific Test Package

```bash
go test ./internal/domain
```

## Project Structure

```
classical-tagger/
├── cmd/
│   ├── validate/     # Validation CLI
│   ├── tag/          # Tagging CLI
│   └── extract/      # Extraction CLI
├── internal/
│   ├── domain/       # Domain models (Album, Track, Artist, etc.)
│   ├── tagging/      # Tag reading/writing
│   ├── validation/   # Validation rules engine
│   ├── scraping/     # Web scraping
│   └── storage/      # JSON persistence
└── go.mod
```

## Domain Model

### Value Objects (Implemented)

- **Level**: ERROR, WARNING, INFO
- **Role**: Composer, Soloist, Ensemble, Conductor, Arranger, Guest
- **Artist**: Immutable name + role
- **ValidationIssue**: level + track + rule + message

### Entities (TODO)

- **Track**: disc, track, title, artists[], name
- **Edition**: label, catalogNumber, editionYear
- **Album**: title, originalYear, edition, tracks[]

## Validation Rules

Based on:
1. "Guide for Classical Music uploads :: Redacted"
2. "Name formatting" rules (higher priority)

Key rules:
- Mandatory tags: Composer, Artist Name, Track Title, Album Title, Track Number
- Composer NOT in track title tag
- 180 character path limit
- Specific filename format: `## - Track Title.flac`
- Multi-disc subdirectories

## Development

This project follows Test-Driven Development (TDD) principles. All new features should:
1. Start with failing tests
2. Implement minimal code to pass tests
3. Refactor while keeping tests green

## License

MIT
