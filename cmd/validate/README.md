# validate - Classical Music Directory Validator

Validates classical music album directories for compliance with torrent site rules.

## Features

- ✅ **Directory structure validation** - Multi-disc detection, path length limits
- ✅ **FLAC tag validation** - Required tags, composer rules, artist format
- ✅ **Filename validation** - Track numbering, title format
- ✅ **Rule references** - Each issue includes rule section numbers
- ✅ **Colored output** - Visual indicators for errors, warnings, and info

## Installation

```bash
cd cmd/validate
go build -o validate
```

## Usage

```bash
# Validate a single album directory
./validate "/path/to/Bach - Goldberg Variations (1981) - FLAC"

# The command will scan the directory recursively and report all issues
```

## Output Example

```
=== Validation Report: /music/Bach - Goldberg Variations (1981) - FLAC ===

📁 DIRECTORY STRUCTURE ISSUES:
⚠️  WARNING [Directory] [2.3.2] Folder name should contain the album title

🏷️  METADATA ISSUES:
❌ ERROR [Track 1] [classical.composer] Composer surname "Bach" found in track title
⚠️  WARNING [Album] [2.3.16.4] Edition information recommended

=== SUMMARY ===
❌ FAILED: Album has critical errors
  Structure issues: 1
  Metadata issues: 2
  Read errors: 0
```

## Exit Codes

- `0` - Success (no errors)
- `1` - Validation errors found or invalid arguments

## Validation Levels

- **ERROR** (❌) - Critical issues that violate rules
- **WARNING** (⚠️) - Recommended practices not followed
- **INFO** (ℹ️) - Suggestions for improvement

## What It Checks

### Directory Structure
- Path length (180 character limit)
- Leading spaces in paths
- Multi-disc organization (CD1, CD2, etc.)
- Folder naming conventions

### Metadata
- Required tags: Composer, Artist, Album, Title, Track Number
- Composer NOT in track title
- Artist format validation
- Track number format
- Album completeness

### File Names
- Track number format (01, 02, etc.)
- File extension (.flac)
- Character encoding

## Dependencies

- `github.com/dhowden/tag` - FLAC tag reading
- `github.com/cehbz/classical-tagger/internal/domain`
- `github.com/cehbz/classical-tagger/internal/validation`
- `github.com/cehbz/classical-tagger/internal/filesystem`
- `github.com/cehbz/classical-tagger/internal/tagging`

## Testing

```bash
go test ./cmd/validate -v
```

## Integration with CI/CD

```bash
# Exit code can be used in scripts
if ./validate "/path/to/album"; then
    echo "Album is valid"
else
    echo "Album has errors"
    exit 1
fi
```

## Known Limitations

1. **Artist parsing**: Currently treats entire Artist tag as ensemble
   - TODO: Parse "Soloist, Ensemble, Conductor" format
2. **Arranger detection**: "(arr. by X)" not auto-parsed yet
3. **Title case**: Capitalization not validated yet
4. **Movement format**: Opus/movement numbers not validated yet

## Future Enhancements

- [ ] JSON output format for automation
- [ ] Batch validation of multiple directories
- [ ] Configurable rule severity levels
- [ ] Auto-fix suggestions
- [ ] Integration with `tag` command for fixes
