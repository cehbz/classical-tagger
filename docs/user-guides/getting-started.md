# Getting Started Guide

This guide walks you through your first complete workflow with classical-tagger.

## Installation

### 1. Install Prerequisites

**Go 1.25+**
```bash
# Check your Go version
go version

# If needed, download from https://go.dev/dl/
```

**mktorrent** (for upload command)
```bash
# Ubuntu/Debian
sudo apt-get install mktorrent

# macOS
brew install mktorrent

# Verify
mktorrent -h
```

### 2. Build the Tools

```bash
# Clone repository
git clone https://github.com/cehbz/classical-tagger
cd classical-tagger

# Build all commands
go build -o validate cmd/validate/main.go
go build -o extract cmd/extract/main.go
go build -o tag cmd/tag/main.go
go build -o upload cmd/upload/main.go

# Optional: Install to PATH
sudo cp validate extract tag upload /usr/local/bin/
```

### 3. Set Up Configuration

Create `~/.config/classical-tagger/config.yaml`:

```bash
mkdir -p ~/.config/classical-tagger
```

Edit `~/.config/classical-tagger/config.yaml`:

```yaml
# Discogs API token
# Get yours at: https://www.discogs.com/settings/developers
discogs:
  token: "your-discogs-token-here"

# Redacted API key
# Get yours at: https://redacted.sh/user.php?action=edit (Access Settings)
redacted:
  api_key: "your-redacted-api-key-here"

# Optional: Cache settings
cache:
  ttl_hours: 24  # Default: 24
```

**Getting API Keys:**

**Discogs:**
1. Go to https://www.discogs.com/settings/developers
2. Click "Generate new token"
3. Copy the token to your config file

**Redacted:**
1. Go to https://redacted.sh/user.php?action=edit
2. Navigate to "Access Settings" → "API Keys"
3. Create new key with "Torrents" and "Upload" permissions
4. Copy the API key to your config file

## Your First Workflow

Let's walk through a complete example: fixing a classical torrent with bad tags.

### Scenario

You've downloaded torrent ID `123456` from Redacted. It's a Bach Goldberg Variations recording, but:
- Composer names are inconsistent ("J.S. Bach" vs "Johann Sebastian Bach")
- Track titles don't follow classical format
- Some performer credits are missing

You found the correct metadata on Discogs (release ID `1234567`).

### Step 1: Validate the Downloaded Torrent

First, let's see what's wrong:

```bash
cd ~/Downloads/Bach-Goldberg-Variations-[Redacted]
validate --dir .
```

**Output:**
```
🔍 VALIDATING TORRENT

=== VALIDATION ISSUES ===
❌ [ERROR] Track 1 [classical.composer] Composer name must not appear in track title
❌ [ERROR] Track 2 [classical.composer] Composer name must not appear in track title
⚠️  [WARNING] Track 5 [classical.artist-format] Artist "Johann Sebastian Bach" should be "J.S. Bach"
...

=== SUMMARY ===
❌ FAILED: Album has critical errors
  Issues: 15
```

### Step 2: Extract Correct Metadata

Get the proper metadata from Discogs:

```bash
extract --url "https://www.discogs.com/release/1234567" --output ~/goldberg-metadata.json
```

**Output:**
```
Extracting metadata from: https://www.discogs.com/release/1234567
✓ Fetched HTML (3.2 KB)
✓ Detected source: Discogs
✓ Extracted metadata
  Album: Goldberg Variations, BWV 988
  Tracks: 32
  Composer: Johann Sebastian Bach
  Performer: Glenn Gould (piano)

Writing to: ~/goldberg-metadata.json
✓ Metadata saved
```

**Check the output:**
```bash
cat ~/goldberg-metadata.json
```

You should see properly formatted JSON with all tracks and metadata.

### Step 3: Fix the Tags

Apply the correct metadata:

```bash
# First, dry run to see what will happen
tag --metadata ~/goldberg-metadata.json \
    --dir ~/Downloads/Bach-Goldberg-Variations-[Redacted] \
    --dry-run \
    --verbose
```

**Output:**
```
Loading metadata from ~/goldberg-metadata.json...
✓ Loaded album: Goldberg Variations, BWV 988 (1981)
  Tracks: 32

Validating metadata...
✓ Metadata is valid

Scanning directory: ~/Downloads/Bach-Goldberg-Variations-[Redacted]
✓ Found 32 FLAC files

Matching tracks to files...
✓ Track 1 → 01 Aria.flac
✓ Track 2 → 02 Variation 1.flac
...

[DRY RUN] Would write tagged files to: ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged

=== Summary ===
Would update: 32 files
```

**Looks good! Now do it for real:**

```bash
tag --metadata ~/goldberg-metadata.json \
    --dir ~/Downloads/Bach-Goldberg-Variations-[Redacted]
```

**Output:**
```
...
Writing tagged files to: ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged
✓ Updated 01 Aria.flac
✓ Updated 02 Variation 1.flac
...

=== Summary ===
✓ Successfully updated: 32 files

📁 Tagged files written to: ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged
```

### Step 4: Validate the Fixed Files

Make sure everything is correct:

```bash
validate --dir ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged
```

**Output:**
```
🔍 VALIDATING TORRENT

✅ No validation issues found

=== SUMMARY ===
✅ PASSED: Album is fully compliant
  Issues: 0
```

### Step 5: Upload to Redacted

Now you can trump the original bad torrent:

```bash
# Always dry run first!
upload --dir ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged \
       --torrent 123456 \
       --dry-run \
       --verbose
```

**Output:**
```
[UPLOAD] Starting upload workflow for torrent ID 123456
[UPLOAD] Fetching torrent metadata...
[UPLOAD] Fetching group metadata...
[UPLOAD] Validating artist consistency...
✓ Artists are consistent

[UPLOAD] Merging metadata...
[UPLOAD] Creating torrent file...

=== Upload Metadata ===
Title: Goldberg Variations, BWV 988
Artists: J.S. Bach (composer), Glenn Gould (piano)
Year: 1981
Format: FLAC
Tracks: 32

Trump Reason: Fixed composer name formatting, track titles, and performer credits per classical guidelines

[DRY RUN] Would upload to: https://redacted.sh/upload.php
```

**Everything looks good! Upload for real:**

```bash
upload --dir ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged \
       --torrent 123456
```

**Output:**
```
...
✓ Torrent created: ~/Downloads/Bach-Goldberg-Variations-[Redacted]_tagged.torrent
✓ Upload successful!

New torrent ID: 789012
View at: https://redacted.sh/torrents.php?id=98765&torrentid=789012
```

### Step 6: Verify on Redacted

1. Visit the link in your browser
2. Check that metadata is correct
3. Download and seed the new torrent
4. Original torrent #123456 should now be trumped

## Common Workflows

### Workflow 1: Quick Validation

Just check if a torrent is compliant:

```bash
validate --dir /path/to/torrent
```

### Workflow 2: Extract Without Discogs

Extract metadata from existing FLAC tags:

```bash
extract --dir /path/to/torrent --output metadata.json
```

### Workflow 3: Custom Trump Reason

Be specific about what you fixed:

```bash
upload --dir ./fixed --torrent 123456 \
  --reason "Corrected composer names (Bach -> J.S. Bach), fixed multi-movement work titles, added conductor credit"
```

### Workflow 4: Fix Only Specific Tracks

If you only need to fix a few tracks:

1. Extract metadata: `extract --dir ./torrent --output meta.json`
2. Edit `meta.json` manually to fix specific tracks
3. Apply tags: `tag --metadata meta.json --dir ./torrent`

## Tips and Best Practices

### Before Uploading

✅ **DO:**
- Always run `validate` on your fixed torrent
- Always use `--dry-run` on `upload` first
- Check the extracted metadata JSON before applying
- Keep backups of original torrents (until upload succeeds)
- Be specific in your trump reasons

❌ **DON'T:**
- Upload without validating first
- Skip the dry-run
- Assume extracted metadata is perfect
- Trump for minor cosmetic differences
- Use generic trump reasons

### Working with Metadata

**Checking metadata quality:**
```bash
# Extract and immediately validate
extract --url "..." --output meta.json
validate --metadata meta.json
```

**Comparing metadata sources:**
```bash
# Extract from Discogs
extract --url "https://discogs.com/..." --output discogs.json

# Extract from existing files
extract --dir ./torrent --output local.json

# Validate against reference
validate --metadata local.json --reference discogs.json
```

### Troubleshooting

**"No FLAC files found"**
- Check you're in the right directory
- Ensure files have .flac extension (not .FLAC)

**"Validation errors in metadata"**
- Review the error messages carefully
- Edit the JSON file to fix issues
- Or extract from a different source

**"Artist role mismatch"**
- Redacted has different artist credits than your tags
- Review carefully—you may need to fix your metadata
- Don't blindly override Redacted's data

**"Rate limit exceeded"**
- You're making too many API requests
- Wait 10 seconds and try again
- Consider increasing cache TTL in config

## Next Steps

Now that you've completed your first workflow:

1. Read the individual tool guides:
   - [Validate](validate.md)
   - [Extract](extract.md)
   - [Tag](tag.md)
   - [Upload](upload.md)

2. Learn about the architecture:
   - [Architecture](../../project/ARCHITECTURE.md)

3. Contribute:
   - [Development Guide](../../project/DEVELOPMENT.md)

## Getting Help

- **Troubleshooting:** See [Troubleshooting Guide](troubleshooting.md)
- **Issues:** Report bugs on [GitHub Issues](https://github.com/cehbz/classical-tagger/issues)
- **Questions:** Ask in [GitHub Discussions](https://github.com/cehbz/classical-tagger/discussions)