# Upload Tool User Guide

## Quick Start

### First Time Setup

1. **Get a Redacted API Key**
   ```bash
   # Visit https://redacted.sh/user.php?action=edit
   # Go to "Access Settings" → "API Keys"
   # Create key with "Torrents" and "Upload" permissions
   ```

2. **Set Your API Key**
   ```bash
   echo 'export REDACTED_API_KEY="your-key-here"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Install mktorrent**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install mktorrent
   
   # macOS
   brew install mktorrent
   
   # Verify installation
   mktorrent -h
   ```

### Your First Upload (Trump)

Let's say you downloaded a torrent (ID 123456) with bad tags:

```bash
# Step 1: Extract proper metadata
extract --discogs 11245120 --output metadata.json

# Step 2: Fix the tags
tag --dir ./bad_torrent --reference metadata.json --output ./fixed_torrent

# Step 3: Validate the fixes
validate --dir ./fixed_torrent

# Step 4: Upload (trump) - always dry run first!
upload --dir ./fixed_torrent --torrent 123456 --dry-run --verbose

# Step 5: If everything looks good, do it for real
upload --dir ./fixed_torrent --torrent 123456
```

## Common Scenarios

### Scenario 1: Fix Wrong Composer Names

You found a torrent where Bach is credited as "Johann Sebastian Bach" instead of "J.S. Bach":

```bash
# Get the correct metadata from Discogs
extract --discogs [discogs-id] --output correct_metadata.json

# Fix the tags
tag --dir ./downloaded --reference correct_metadata.json --output ./fixed

# Upload with specific reason
upload --dir ./fixed --torrent 123456 \
  --reason "Standardized composer names per classical guidelines"
```

### Scenario 2: Fix Work Groupings

A Poulenc album has movements not properly grouped:

```bash
# The metadata file should have correct work groupings
extract --url "https://www.discogs.com/..." --output metadata.json

# Tag will fix the track titles to include work names
tag --dir ./downloaded --reference metadata.json --output ./fixed

# Upload
upload --dir ./fixed --torrent 789012 \
  --reason "Fixed multi-movement work titles and groupings"
```

### Scenario 3: Fix Performer Credits

Orchestra and conductor credits are wrong or missing:

```bash
# Dry run first to see what will happen
upload --dir ./fixed --torrent 345678 --dry-run --verbose

# Check the artist validation output carefully
# If there are conflicts, you may need to re-tag

# Once validation passes
upload --dir ./fixed --torrent 345678 \
  --reason "Corrected performer credits and roles"
```

## Understanding the Output

### Dry Run Output

```
[UPLOAD] Starting upload workflow for torrent ID 123456
[UPLOAD] Fetching torrent metadata...
[UPLOAD] Using cached torrent metadata
[UPLOAD] Fetching group metadata for group ID 98765...
[UPLOAD] Validating artist consistency...
[UPLOAD] Merging metadata...
[UPLOAD] Creating torrent file...

=== Upload Metadata ===
Title: Noël! Christmas! Weihnachten!
Year: 2013
Format: FLAC / Lossless / CD
Label: Harmonia Mundi - HMC 902170

Artists:
  - RIAS Kammerchor

Composers:
  - Felix Mendelssohn
  - Johannes Brahms
  ...

Conductors:
  - Hans-Christoph Rademann

Tags: classical, choral, sacred

Trump Reason: Corrected tags and filenames according to classical music guidelines

Description:
[Original description preserved...]

[Trump Upload] Fixed: Corrected tags and filenames according to classical music guidelines
```

### Validation Errors

If you see validation errors:

```
Validation error: artist "Hans-Christoph Rademann" role mismatch: 
  Redacted has "conductor", local has "composer"
```

This means your tags don't match what's on Redacted. You need to either:
1. Fix your tags to match Redacted's roles
2. Report the issue if Redacted is wrong

### Success Message

```
Upload completed successfully!
```

Your torrent has been uploaded and the original has been trumped.

## Tips and Tricks

### 1. Always Use References

Don't try to fix tags manually. Use metadata from:
- Discogs (most reliable for classical)
- Presto Classical
- Original label websites

### 2. Check Your Work

Before uploading:
```bash
# Validate thoroughly
validate --dir ./fixed --verbose

# Compare with original
ls -la ./downloaded/*.flac
ls -la ./fixed/*.flac
```

### 3. Use the Cache

The tool caches API responses for 24 hours:
```bash
# First run fetches from API
upload --dir ./fixed --torrent 123456 --dry-run

# Second run uses cache (faster)
upload --dir ./fixed --torrent 123456 --dry-run

# Force fresh fetch if needed
upload --dir ./fixed --torrent 123456 --clear-cache
```

### 4. Write Good Trump Reasons

Be specific about what you fixed:

**Good:**
- "Fixed composer names to standard format (J.S. Bach not Johann Sebastian Bach)"
- "Corrected multi-movement work titles per classical guidelines"
- "Added missing conductor and orchestra credits"

**Bad:**
- "Fixed tags"
- "Better metadata"
- "Correct version"

### 5. Handle Failures

If upload fails:

```bash
# Check your API key
echo $REDACTED_API_KEY

# Check rate limiting (wait 10 seconds between attempts)
sleep 10
upload --dir ./fixed --torrent 123456

# Verify torrent ID is correct
# Go to https://redacted.sh/torrents.php?id=123456

# Clear cache if metadata seems wrong
upload --dir ./fixed --torrent 123456 --clear-cache
```

## Frequently Asked Questions

### Q: Can I upload new torrents (not trumps)?
A: Not yet. This tool is currently designed specifically for trumping existing torrents.

### Q: What if artist validation fails?
A: The tool is strict about artist consistency. If Redacted has an artist as "conductor" and your tags have them as "composer", you need to fix your tags or determine if Redacted is wrong.

### Q: How long does cache last?
A: 24 hours. Use `--clear-cache` to force refresh.

### Q: Can I upload multiple albums at once?
A: No, process one album at a time for safety.

### Q: What if I uploaded the wrong thing?
A: Contact Redacted staff immediately. The upload tool doesn't have an undo feature.

### Q: Why is it failing with "mktorrent not found"?
A: Install mktorrent: `sudo apt-get install mktorrent` (Ubuntu/Debian) or `brew install mktorrent` (macOS)

### Q: Can I use this for non-classical music?
A: The tool is optimized for classical metadata. It may work for other genres but hasn't been tested.

## Safety Checklist

Before every upload:

- [ ] Run `validate` on your fixed files
- [ ] Do a `--dry-run` first
- [ ] Check artist validation passed
- [ ] Verify the torrent ID is correct
- [ ] Write a clear trump reason
- [ ] Make sure you're in the right directory

## Getting Help

1. **Run with verbose mode**: `--verbose` shows detailed progress
2. **Check the README**: More technical details available
3