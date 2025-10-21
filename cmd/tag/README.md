# tag - Apply Metadata to FLAC Files

Applies JSON metadata to FLAC audio files.

## Features

- ✅ **Load metadata from JSON** - Uses the same format as storage package
- ✅ **Automatic file matching** - Matches tracks to files by name
- ✅ **Validation** - Validates metadata before applying (unless --force)
- ✅ **Backup system** - Creates timestamped backups before modification
- ✅ **Dry-run mode** - Preview changes without modifying files
- ✅ **Error recovery** - Restores from backup if write fails

## Installation

```bash
cd cmd/tag
go build -o tag
```

## Usage

### Basic Usage

```bash
# Apply metadata to files in current directory
./tag -metadata album.json

# Apply to specific directory
./tag -metadata album.json -dir /path/to/album

# Dry run (preview without changes)
./tag -metadata album.json -dry-run

# Skip validation
./tag -metadata album.json -force

# Disable backups (not recommended)
./tag -metadata album.json -backup=false
```

### Complete Example

```bash
# 1. Create or obtain metadata JSON
cat > metadata.json << 'EOF'
{
  "title": "Goldberg Variations",
  "original_year": 1981,
  "edition": {
    "label": "Sony Classical",
    "catalog_number": "SMK89245",
    "edition_year": 1981
  },
  "tracks": [
    {
      "disc": 1,
      "track": 1,
      "title": "Aria",
      "composer": {
        "name": "Johann Sebastian Bach",
        "role": "composer"
      },
      "artists": [
        {
          "name": "Glenn Gould",
          "role": "soloist"
        }
      ],
      "name": "01 Aria.flac"
    }
  ]
}
EOF

# 2. Dry run to preview changes
./tag -metadata metadata.json -dry-run

# 3. Apply tags
./tag -metadata metadata.json

# 4. Verify with validate CLI
cd ../validate
./validate /path/to/album
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-metadata` | *required* | Path to metadata JSON file |
| `-dir` | `.` | Target directory containing FLAC files |
| `-dry-run` | `false` | Show what would be done without doing it |
| `-backup` | `true` | Create backup before modifying files |
| `-force` | `false` | Skip validation and apply anyway |

## Output Example

```
Loading metadata from album.json...
✓ Loaded album: Goldberg Variations (1981)
  Tracks: 32

Validating metadata...
✓ Metadata is valid

Scanning directory: /music/Bach - Goldberg Variations
✓ Found 32 FLAC files

Matching tracks to files...
✓ Track 1 -> 01 Aria.flac
✓ Track 2 -> 02 Variation 1.flac
✓ Track 3 -> 03 Variation 2.flac
...

Applying tags...
✓ Updated 01 Aria.flac
✓ Updated 02 Variation 1.flac
✓ Updated 03 Variation 2.flac
...

=== Summary ===
✓ Successfully updated: 32 files

💾 Backups created:
  /music/Bach - Goldberg Variations/01 Aria.flac.20250120-143022.bak
  ...
```

## Error Handling

### Validation Errors

```
Validating metadata...
❌ [ERROR] Track 1 [classical.composer] Composer name must not appear in track title

❌ Metadata has errors. Fix them or use --force to proceed anyway.
```

### File Matching Issues

```
Matching tracks to files...
✓ Track 1 -> 01 Aria.flac
⚠️  No file found for track 2: Variation 1

⚠️  1 tracks could not be matched to files
Use --force to proceed anyway
```

### Write Failures

If writing tags fails, the backup is automatically restored:

```
❌ Failed to write tags to 01 Aria.flac: permission denied
   Restored from backup
```

## Backup System

Backups are created with timestamps:
```
original-file.flac          # Original
original-file.flac.20250120-143022.bak  # Backup
```

To restore manually:
```bash
cp file.flac.20250120-143022.bak file.flac
```

## Important Notes

### ⚠️ Current Limitation

**FLAC tag writing is not yet fully implemented**. The current implementation provides:
- ✅ Complete interface and CLI
- ✅ File matching and validation
- ✅ Backup/restore system
- ✅ Dry-run mode
- ❌ Actual tag writing (returns "not yet implemented")

To complete the implementation, we need to add a FLAC tag writing library. Options:

1. **metaflac command-line tool** (simplest)
   ```go
   cmd := exec.Command("metaflac",
       "--remove-all-tags",
       "--set-tag=TITLE="+title,
       "--set-tag=COMPOSER="+composer,
       path)
   ```

2. **go-flac library** - Pure Go implementation
3. **Implement vorbis comment writing** - From scratch

For now, use `--dry-run` to test the workflow.

## Integration with Other Commands

### Workflow

```bash
# 1. Extract metadata from web
cd ../extract
./extract -url https://www.harmoniamundi.com/... -output album.json

# 2. Validate before applying
cd ../validate
./validate /path/to/album

# 3. Apply metadata
cd ../tag
./tag -metadata album.json -dir /path/to/album

# 4. Verify it worked
cd ../validate
./validate /path/to/album
```

## Testing

```bash
# Run tests
go test -v

# Test with dry-run
./tag -metadata test.json -dry-run

# Test backup system
./tag -metadata test.json -dir testdata
# Check that .bak files were created
```

## Troubleshooting

### "No FLAC files found"

- Check the directory path
- Ensure files have .flac extension
- Check file permissions

### "No file found for track X"

- Ensure track `name` field matches actual filename
- Try updating JSON with correct filenames
- Use --force to proceed with partial matches

### "Metadata has errors"

- Run validate CLI to see detailed issues
- Fix issues in JSON file
- Or use --force to apply anyway (not recommended)

## Future Enhancements

- [ ] Implement actual FLAC tag writing
- [ ] Support for other audio formats (MP3, M4A)
- [ ] Fuzzy matching for filenames
- [ ] Batch processing multiple albums
- [ ] Progress bars for large albums
- [ ] Parallel processing
- [ ] Undo command
- [ ] Template-based filename generation

## Safety

The tag CLI is designed to be safe:
- ✅ Validates before applying
- ✅ Creates backups by default
- ✅ Restores on error
- ✅ Dry-run mode for testing
- ✅ Clear error messages

Always test with `--dry-run` first!
