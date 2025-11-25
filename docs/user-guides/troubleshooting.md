# Troubleshooting Guide

This guide covers common errors, problems, and solutions when using classical-tagger.

## Table of Contents

- [Configuration Issues](#configuration-issues)
- [Validation Errors](#validation-errors)
- [Extraction Problems](#extraction-problems)
- [Tagging Issues](#tagging-issues)
- [Upload Failures](#upload-failures)
- [General Problems](#general-problems)

---

## Configuration Issues

### "Config file not found"

**Error:**
```
Error loading API key from config: config file not found at ~/.config/classical-tagger/config.yaml: please create it with your Discogs token
```

**Cause:** Configuration file doesn't exist.

**Solution:**
```bash
# Create config directory
mkdir -p ~/.config/classical-tagger

# Create config file
cat > ~/.config/classical-tagger/config.yaml << 'EOF'
discogs:
  token: "your-discogs-token-here"

redacted:
  api_key: "your-redacted-api-key-here"

cache:
  ttl_hours: 24
EOF

# Edit with your actual tokens
nano ~/.config/classical-tagger/config.yaml
```

**Get API Keys:**
- Discogs: https://www.discogs.com/settings/developers
- Redacted: https://redacted.sh/user.php?action=edit (Access Settings)

---

### "Discogs token not found in config file"

**Error:**
```
Discogs token not found in config file: please add 'discogs.token' to ~/.config/classical-tagger/config.yaml
```

**Cause:** Config file exists but missing `discogs.token` field.

**Solution:**
Add to your config file:
```yaml
discogs:
  token: "your-actual-token-here"
```

---

### "Redacted API key not found in config file"

**Error:**
```
Redacted API key not found in config file: please add 'redacted.api_key' to ~/.config/classical-tagger/config.yaml
```

**Cause:** Config file exists but missing `redacted.api_key` field.

**Solution:**
Add to your config file:
```yaml
redacted:
  api_key: "your-actual-key-here"
```

---

### "Failed to parse config file"

**Error:**
```
failed to parse config file: yaml: line 3: found character that cannot start any token
```

**Cause:** Invalid YAML syntax in config file.

**Solution:**
Check your YAML syntax:
- Proper indentation (2 spaces, not tabs)
- No special characters in strings (use quotes)
- Colons followed by space

**Valid format:**
```yaml
discogs:
  token: "my-token-123"

redacted:
  api_key: "my-key-456"
```

**Invalid format:**
```yaml
discogs:
token:"my-token-123"  # Missing space after colon, wrong indentation
```

---

## Validation Errors

### "Composer name must not appear in track title"

**Error:**
```
❌ [ERROR] Track 1 [classical.composer] Composer name must not appear in track title
```

**Cause:** Track title contains the composer's name (e.g., "Bach - Aria").

**Solution:**
The composer belongs in the COMPOSER tag, not the track title:

**Wrong:**
```
TITLE: Bach - Aria
COMPOSER: Johann Sebastian Bach
```

**Right:**
```
TITLE: Aria
COMPOSER: Johann Sebastian Bach
```

**Fix it:**
```bash
# Extract correct metadata
extract --url "..." --output meta.json

# Edit meta.json if needed
# Apply tags
tag --metadata meta.json --dir ./album
```

---

### "Artist name format incorrect"

**Error:**
```
⚠️  [WARNING] Track 5 [classical.artist-format] Artist "Johann Sebastian Bach" should be "J.S. Bach"
```

**Cause:** Composer name doesn't follow standardized format.

**Solution:**
Use abbreviated forms per classical music guidelines:
- "J.S. Bach" not "Johann Sebastian Bach"
- "W.A. Mozart" not "Wolfgang Amadeus Mozart"
- "L. van Beethoven" not "Ludwig van Beethoven"

**Fix it:** Extract from authoritative source or manually edit metadata JSON.

---

### "Path exceeds 180 character limit"

**Error:**
```
❌ [ERROR] Track 3 [filename.path-length] File path exceeds 180 character limit (current: 195)
```

**Cause:** Full path to file is too long for some filesystems.

**Solution:**
Shorten the path by:
1. Using shorter root directory names
2. Abbreviating work titles
3. Removing unnecessary information from titles

Example:
```
# Too long (195 chars)
/music/Johann-Sebastian-Bach-Goldberg-Variations-BWV-988-Glenn-Gould-Piano-1981-Sony-Classical/CD1/01 - Aria from Goldberg Variations BWV 988.flac

# Better (145 chars)
/music/Bach-Goldberg-Variations-Gould-1981/CD1/01 - Aria.flac
```

---

### "Disc number mismatch"

**Error:**
```
❌ [ERROR] Track 5 [multi-disc.numbering] Track reports disc 1 but file is in disc 2 subdirectory
```

**Cause:** Track metadata doesn't match directory structure.

**Solution:**
Ensure disc numbers in tags match subdirectory:
- Files in `Disc 1/` should have DISCNUMBER=1
- Files in `Disc 2/` should have DISCNUMBER=2

**Fix it:** Re-tag with correct disc numbers or reorganize directories.

---

## Extraction Problems

### "No FLAC files found in directory"

**Error:**
```
Error: no FLAC files found in directory
```

**Cause:** 
- Directory is empty
- Files have wrong extension (.FLAC instead of .flac)
- Looking in wrong directory

**Solution:**
```bash
# Check directory contents
ls -la /path/to/directory

# Check for uppercase extensions
ls -la *.FLAC

# If files have uppercase extensions, rename them
for f in *.FLAC; do mv "$f" "${f%.FLAC}.flac"; done
```

---

### "HTTP 403 Forbidden" or "HTTP 404 Not Found"

**Error:**
```
Error fetching URL: HTTP 403 Forbidden
```

**Cause:**
- URL is incorrect
- Website is blocking automated requests
- Page doesn't exist

**Solution:**
1. Verify URL in browser first
2. Copy exact URL from address bar
3. Try different metadata source
4. Check if site is blocking bots (use browser developer tools)

---

### "Extraction failed due to required field errors"

**Error:**
```
Extraction failed: missing required fields: title, year
```

**Cause:** Critical metadata couldn't be extracted from website.

**Solution:**
```bash
# Use verbose mode to see what was found
extract --url "..." --output meta.json --verbose

# Try different source
extract --url "different-site-url" --output meta.json

# Use --force to create output anyway (not recommended for tagging)
extract --url "..." --output meta.json --force

# Or extract from existing FLAC files
extract --dir ./album --output meta.json
```

---

### "Domain conversion failed"

**Error:**
```
Domain conversion failed: validation errors in extracted metadata
```

**Cause:** Extracted metadata doesn't meet validation rules.

**Solution:**
```bash
# Check what was extracted
cat meta.json

# Validate to see specific errors
validate --metadata meta.json

# Edit JSON manually to fix issues
nano meta.json

# Or use --force to skip validation (not recommended)
extract --url "..." --output meta.json --force
```

---

### "Rate limit exceeded"

**Error:**
```
HTTP 429: Rate limit exceeded. Retry-After: 10
```

**Cause:** Too many API requests in short time.

**Solution:**
- Wait 10 seconds and try again
- Check cache directory: `~/.cache/classical-tagger/`
- Increase cache TTL in config:
  ```yaml
  cache:
    ttl_hours: 48  # Default is 24
  ```
- Use `--clear-cache` sparingly

---

## Tagging Issues

### "No file found for track X"

**Error:**
```
⚠️  No file found for track 2: Variation 1
```

**Cause:** 
- Track numbers in JSON don't match filename prefixes
- Missing files
- Incorrect filename format

**Solution:**
Check filename prefixes match track numbers:

**JSON:**
```json
{
  "disc": 1,
  "track": 2,
  "title": "Variation 1"
}
```

**Filename must start with:** `02` (or `02-`, `02.`, `02_`, `02 `)

**Valid filenames:**
- `02 - Variation 1.flac`
- `02-Variation 1.flac`
- `02. Variation 1.flac`

**Invalid filenames:**
- `2 - Variation 1.flac` (single digit)
- `Track 2 - Variation 1.flac` (no leading number)

**Fix it:**
```bash
# Rename files with single-digit track numbers
for f in [0-9]\ -*.flac; do 
  mv "$f" "0$f"
done
```

---

### "Metadata has errors"

**Error:**
```
❌ Metadata has errors. Fix them or use --force to proceed anyway.
```

**Cause:** Metadata JSON file has validation errors.

**Solution:**
```bash
# See what's wrong
validate --metadata album.json

# Fix the issues in the JSON file
nano album.json

# Try again
tag --metadata album.json --dir ./album

# Only use --force if you understand the risks
tag --metadata album.json --dir ./album --force
```

---

### "Permission denied"

**Error:**
```
Error: failed to write file: permission denied
```

**Cause:**
- Output directory is read-only
- Insufficient permissions
- Disk full

**Solution:**
```bash
# Check permissions
ls -la /path/to/output

# Check disk space
df -h

# Specify different output directory
tag --metadata meta.json --dir ./source --output ~/tagged

# Fix permissions if needed
chmod -R u+w /path/to/output
```

---

### "Tag writing not yet implemented"

**Error:**
```
Error: FLAC tag writing not yet implemented
```

**Cause:** FLAC writing library needs testing with real files.

**Solution:**
This is a known limitation. The interface is complete but needs testing:

```bash
# For now, use dry-run to verify workflow
tag --metadata meta.json --dir ./album --dry-run

# Or use external tools temporarily
metaflac --set-tag="TITLE=..." file.flac
```

**Track issue:** Check project GitHub for FLAC writing implementation status.

---

## Upload Failures

### "Torrent ID not found"

**Error:**
```
Error: API error 404: Torrent not found
```

**Cause:** 
- Torrent ID is incorrect
- Torrent was deleted
- You don't have access to this torrent

**Solution:**
- Verify torrent ID on Redacted website
- Check if torrent still exists
- Ensure you have permission to view it

---

### "Artist role mismatch"

**Error:**
```
Validation error: artist "Glenn Gould" role mismatch: Redacted has "soloist", local has "conductor"
```

**Cause:** Your tags have different artist roles than Redacted's database.

**Solution:**
This is a safety check to prevent incorrect uploads:

1. **Check which is correct:**
   - Visit torrent page on Redacted
   - Check original artist credits
   - Verify against liner notes

2. **If your tags are wrong:**
   ```bash
   # Re-extract or manually fix metadata
   nano metadata.json
   # Re-tag
   tag --metadata metadata.json --dir ./album
   ```

3. **If Redacted is wrong:**
   - This might not be trumpable just for artist roles
   - Consider editing group info on Redacted instead

---

### "Missing required fields: tags, year"

**Error:**
```
Error: missing required fields: tags, year
```

**Cause:** Metadata extraction didn't capture all required fields.

**Solution:**
```bash
# Check what's missing
cat metadata.json

# Extract from better source
extract --url "better-source-url" --output meta.json

# Or add manually
nano metadata.json
```

Required fields:
- `title` - Album title
- `original_year` - Original release year
- `tracks` - Array of tracks with metadata

---

### "API authentication failed"

**Error:**
```
Error: API error 401: Invalid API key
```

**Cause:**
- API key is incorrect
- API key expired
- API key lacks required permissions

**Solution:**
```bash
# Check your config file
cat ~/.config/classical-tagger/config.yaml

# Get new API key from Redacted:
# https://redacted.sh/user.php?action=edit
# Ensure it has "Torrents" and "Upload" permissions

# Update config file
nano ~/.config/classical-tagger/config.yaml

# Or override temporarily
upload --api-key "your-key" --dir ./album --torrent 123456
```

---

### "mktorrent not found"

**Error:**
```
Error: mktorrent executable not found in PATH
```

**Cause:** mktorrent is not installed.

**Solution:**
```bash
# Ubuntu/Debian
sudo apt-get install mktorrent

# macOS
brew install mktorrent

# Verify
mktorrent -h
```

---

## General Problems

### "Command not found"

**Error:**
```
bash: validate: command not found
```

**Cause:** 
- Tool not built yet
- Tool not in PATH
- Wrong directory

**Solution:**
```bash
# Build the tool
cd classical-tagger
go build -o validate cmd/validate/main.go

# Run from current directory
./validate --dir ./album

# Or install to PATH
sudo cp validate /usr/local/bin/

# Or add to PATH temporarily
export PATH=$PATH:$(pwd)
```

---

### "Go version too old"

**Error:**
```
go: go.mod requires go >= 1.25
```

**Cause:** Your Go version is older than 1.25.

**Solution:**
```bash
# Check version
go version

# Upgrade Go
# Download from https://go.dev/dl/

# Or use version manager
# https://github.com/moovweb/gvm
```

---

### "Cannot connect to cache directory"

**Error:**
```
Warning: failed to clear cache: permission denied
```

**Cause:** No write permission to cache directory.

**Solution:**
```bash
# Check cache directory
ls -la ~/.cache/classical-tagger

# Fix permissions
chmod -R u+w ~/.cache/classical-tagger

# Or specify different cache directory
export XDG_CACHE_HOME=/path/to/cache
```

---

### "Slow extraction/upload"

**Problem:** Operations take a long time.

**Causes & Solutions:**

**Slow network:**
```bash
# Increase timeout
extract --url "..." --output meta.json --timeout 120s
```

**Rate limiting:**
```bash
# Check if you're being rate-limited
# Wait and try again
# Check cache is working

# Verify cache directory
ls -la ~/.cache/classical-tagger
```

**Large torrent:**
```bash
# This is normal for 100+ tracks
# Use --verbose to see progress
upload --dir ./large-album --torrent 123456 --verbose
```

---

### "UTF-8 encoding issues"

**Problem:** Special characters appear as `�` or garbled.

**Examples:**
- é → Ã©
- ö → Ã¶
- ñ → Ã±

**Solution:**
This should be handled automatically. If you see issues:

```bash
# Check file encoding
file -i file.flac

# Try re-extracting
extract --url "..." --output meta.json --force

# If still broken, file a bug report with:
# - URL that was extracted
# - Expected vs actual characters
```

---

## Debugging Tips

### Enable Verbose Output

```bash
# See detailed progress
validate --dir ./album --verbose
extract --url "..." --output meta.json --verbose
tag --metadata meta.json --dir ./album --verbose
upload --dir ./album --torrent 123456 --verbose
```

### Use Dry Run

```bash
# Test without making changes
tag --metadata meta.json --dir ./album --dry-run
upload --dir ./album --torrent 123456 --dry-run
```

### Check Logs

```bash
# Capture output to file
validate --dir ./album > validate.log 2>&1

# View log
cat validate.log
```

### Validate at Each Step

```bash
# 1. Validate downloaded torrent
validate --dir ./downloaded

# 2. Validate extracted metadata
validate --metadata meta.json

# 3. Validate after tagging
validate --dir ./tagged

# 4. Compare with reference
validate --metadata ./tagged/meta.json --reference ./original/meta.json
```

### Clear Cache When Needed

```bash
# Clear all cache
rm -rf ~/.cache/classical-tagger

# Or use built-in clear
upload --clear-cache --dir ./album --torrent 123456
```

---

## Getting Help

If you're still stuck:

1. **Check documentation:**
   - [Getting Started Guide](getting-started.md)
   - [Tool-specific guides](validate.md)
   - [Architecture](../../project/ARCHITECTURE.md)

2. **Search existing issues:**
   - GitHub Issues: https://github.com/cehbz/classical-tagger/issues

3. **Ask for help:**
   - GitHub Discussions: https://github.com/cehbz/classical-tagger/discussions
   - Provide: error message, command you ran, relevant config

4. **Report bugs:**
   - Include: full error output, steps to reproduce, OS/Go version
   - Use `--verbose` flag and include output

---

## Common Error Patterns

### "It worked yesterday, now it doesn't"

**Likely causes:**
- Website changed structure (extractors need updating)
- API rate limits reached (wait and retry)
- Cache is stale (clear cache)
- Configuration was modified (check config file)

### "Works on one album, fails on another"

**Likely causes:**
- Different metadata structure (some fields missing)
- Special characters in filenames (check for invalid chars)
- Different disc structure (multi-disc vs single)
- Edge case in validation rules (report as bug)

### "Dry run works, real run fails"

**Likely causes:**
- File permission issues (check write permissions)
- Disk space (check with `df -h`)
- Race condition (file in use, antivirus scanning)
- Network timeout (increase timeout value)

---

**Last Updated:** 2025-01-XX  
**Version:** 1.0  
**Maintainer:** classical-tagger project