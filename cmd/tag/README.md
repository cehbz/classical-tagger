# Tag CLI - Apply Metadata to FLAC Files

## Overview

The `tag` CLI reads metadata from a JSON file and applies it to FLAC files, writing the tagged files to a new output directory. The original files remain untouched.

## Usage

```bash
# Basic usage
tag -metadata album.json -dir /path/to/album

# Specify output directory
tag -metadata album.json -dir /path/to/album -output /path/to/output

# Dry run (show what would be done)
tag -metadata album.json -dir /path/to/album -dry-run

# Skip validation (not recommended)
tag -metadata album.json -dir /path/to/album -force
```

## Flags

- `-metadata FILE` (required) - Path to metadata JSON file
- `-dir DIR` - Directory containing source FLAC files (default: current directory)
- `-output DIR` - Output directory for tagged files (default: `<source>_tagged`)
- `-dry-run` - Show what would be done without modifying files
- `-force` - Skip validation and proceed anyway

## Workflow

### 1. Extract Metadata (future)
```bash
# Extract metadata from website
extract -url https://www.harmoniamundi.com/... -output album.json
```

### 2. Validate Metadata
```bash
# Validate before applying
validate -metadata album.json
```

### 3. Apply Tags
```bash
# Apply to directory
tag -metadata album.json -dir /music/album

# Output structure:
# /music/album/           <- original files (untouched)
# /music/album_tagged/    <- tagged files (new)
```

### 4. Verify Results
```bash
# Validate the tagged directory
validate /music/album_tagged
```

## Output

### Successful Run
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

Writing tagged files to: /music/Bach - Goldberg Variations_tagged
✓ Updated 01 Aria.flac
✓ Updated 02 Variation 1.flac
✓ Updated 03 Variation 2.flac
...

=== Summary ===
✓ Successfully updated: 32 files

📁 Tagged files written to: /music/Bach - Goldberg Variations_tagged
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

If writing tags fails, the error is reported but original files remain untouched:

```
❌ Failed to write 01 Aria.flac: permission denied
```

## File Matching

The CLI matches tracks to files by track number prefix:
- Track 1 matches files starting with "01"
- Track 2 matches files starting with "02"
- etc.

Supported filename patterns:
```
01 Aria.flac
01-Aria.flac
01. Aria.flac
01_Aria.flac
```

## Safety Features

### Non-Destructive
- Original files are never modified
- All tagged files written to separate directory
- Easy to compare original vs tagged

### Validation
- Validates metadata before applying (unless `--force`)
- Reports all validation errors and warnings
- Clear error messages

### Dry Run
- Test workflow without making changes
- Shows exactly what would be done
- Verifies file matching

## Directory Structure

### Before
```
/music/Bach - Goldberg Variations/
├── 01 Aria.flac
├── 02 Variation 1.flac
├── 03 Variation 2.flac
└── ...
```

### After
```
/music/Bach - Goldberg Variations/       <- originals untouched
├── 01 Aria.flac
├── 02 Variation 1.flac
└── ...

/music/Bach - Goldberg Variations_tagged/  <- new tagged files
├── 01 Aria.flac
├── 02 Variation 1.flac
└── ...
```

## Testing

```bash
# Run tests
go test -v

# Test with dry-run
./tag -metadata test.json -dir testdata -dry-run

# Test with real files
./tag -metadata test.json -dir testdata
# Verify output directory created
# Compare original vs tagged
```

## Integration with Other Commands

### Complete Workflow

```bash
# 1. Extract metadata from web (future)
extract -url https://www.harmoniamundi.com/... -output album.json

# 2. Validate metadata
validate -metadata album.json

# 3. Apply to source directory
tag -metadata album.json -dir /music/album

# 4. Verify tagged files
validate /music/album_tagged

# 5. If satisfied, replace originals
rm -rf /music/album
mv /music/album_tagged /music/album
```

## Troubleshooting

### "No FLAC files found"

- Check the directory path
- Ensure files have .flac extension
- Check file permissions

### "No file found for track X"

- Check track numbers in JSON match filename prefixes
- Ensure files use 2-digit track numbers (01, 02, not 1, 2)
- Use --force to proceed with partial matches

### "Metadata has errors"

- Run `validate -metadata album.json` to see detailed issues
- Fix issues in JSON file
- Or use --force to apply anyway (not recommended)

### "Permission denied"

- Check write permissions on output directory
- Ensure output directory is not read-only
- Check disk space

## Future Enhancements

- [ ] Implement actual FLAC tag writing (currently returns "not implemented")
- [ ] Support for other audio formats (MP3, M4A)
- [ ] Fuzzy matching for filenames
- [ ] Batch processing multiple albums
- [ ] Progress bars for large albums
- [ ] Parallel processing
- [ ] Template-based filename generation
- [ ] Automatic backup management

## Current Limitations

### ⚠️ FLAC Tag Writing Not Yet Implemented

The FLAC tag writing functionality uses the `go-flac/flac` and `go-flac/flacvorbis` libraries but requires testing with real FLAC files. The interface and CLI are complete and functional.

To complete the implementation:

1. **Ensure dependencies are installed:**
   ```bash
   go get github.com/go-flac/go-flac
   go get github.com/go-flac/flacvorbis
   ```

2. **Test with real FLAC files:**
   ```bash
   # Create test directory
   mkdir testdata
   cp /path/to/test.flac testdata/01-test.flac
   
   # Test dry-run
   ./tag -metadata test.json -dir testdata -dry-run
   
   # Test actual writing
   ./tag -metadata test.json -dir testdata
   
   # Verify output
   metaflac --list testdata_tagged/01-test.flac
   ```

For now, use `--dry-run` to test the workflow without writing files.

## Safety

The tag CLI is designed to be safe:
- ✅ Original files never modified
- ✅ Validates before applying
- ✅ Tagged files written to separate directory
- ✅ Easy rollback (just delete output directory)
- ✅ Dry-run mode for testing
- ✅ Clear error messages

Always test with `--dry-run` first!
