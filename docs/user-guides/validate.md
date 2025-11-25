# validate - Classical Music JSON Metadata Validator

Validates JSON metadata files extracted from classical music albums for compliance with torrent site rules.

## Features

- ✅ **JSON metadata validation** - Validates extracted album metadata against all validation rules
- ✅ **Reference comparison** - Optionally compares against a reference JSON file
- ✅ **Comprehensive rule checking** - All validation rules including structure, metadata, and formatting
- ✅ **Rule references** - Each issue includes rule section numbers
- ✅ **Colored output** - Visual indicators for errors, warnings, and info

## Installation

```bash
cd cmd/validate
go build -o validate
```

## Usage

```bash
# Validate a JSON metadata file
validate album.json

# Validate against a reference JSON file
validate album.json reference.json
```

## Output Example

```
=== Validation Report ===

Metadata file: album.json
Reference file: reference.json

🏷️  VALIDATION ISSUES:
❌ ERROR [Track 1] [classical.composer] Composer surname "Bach" found in track title
⚠️  WARNING [Album] [2.3.16.4] Edition information recommended
❌ ERROR [Track 5] [2.3.18.2] Track 5 title 'SYMPHONY NO. 5': Not Title Case or Casual Title Case

=== SUMMARY ===
❌ FAILED: Album has critical errors
  Issues: 3
  Load errors: 0
```

## Exit Codes

- `0` - Success (no errors)
- `1` - Validation errors found or invalid arguments

## Validation Levels

- **ERROR** (❌) - Critical issues that violate rules
- **WARNING** (⚠️) - Recommended practices not followed
- **INFO** (ℹ️) - Suggestions for improvement

## What It Checks

The validator checks all validation rules including:

### Metadata Rules
- Required tags: Composer, Artist, Album, Title, Track Number
- Composer NOT in track title
- Artist format validation
- Track number format
- Album completeness
- Tag capitalization (Title Case)

### Structure Rules
- Path length (180 character limit)
- Leading spaces in paths/filenames
- Folder naming conventions
- Multi-disc organization
- Filename format and capitalization

### Reference Comparison
When a reference JSON file is provided, additional checks:
- Tag accuracy vs reference
- Capitalization matching
- Structure consistency

## Dependencies

- `github.com/cehbz/classical-tagger/internal/domain`
- `github.com/cehbz/classical-tagger/internal/validation`
- `github.com/cehbz/classical-tagger/internal/storage`

## Testing

```bash
go test ./cmd/validate -v
```

## Integration with CI/CD

```bash
# Exit code can be used in scripts
if validate album.json; then
    echo "Metadata is valid"
else
    echo "Metadata has errors"
    exit 1
fi
```

## Workflow

1. Extract metadata using `extract` command:
   ```bash
   extract -url "https://..." -output album.json
   ```

2. Validate the extracted JSON:
   ```bash
   validate album.json
   ```

3. Optionally validate against a reference:
   ```bash
   validate album.json reference.json
   ```

4. Apply tags using `tag` command (when implemented)

## Related Commands

- **extract** - Extract metadata from web pages or directories to JSON
- **tag** - Apply metadata to FLAC files (coming soon)
