# Classical Music Tagger - Design Summary

## Project Goals
Build Go applications to:
1. **validate** - Validate tags in downloaded torrent directories
2. **tag** - Apply JSON metadata to audio files
3. **extract** - Extract JSON from classical music web pages

## Key Requirements
- FLAC only (for now)
- Follow "Guide for Classical Music uploads :: Redacted" rules
- Follow "Name formatting" rules (higher priority)
- Classical rules are clarifications of general rules
- Support multiple data sources: Harmonia Mundi, Classical Archives, Naxos, Presto Classical, ArkivMusic
- Validate directory structure and filename conventions (180 char limit)
- TDD with Go 1.25 features

## Domain Model

### Album (Aggregate Root)
- title, originalYear (required)
- edition (optional: label, catalogNumber, editionYear)
- tracks[]
- Validate() returns ALL issues (album + tracks)

### Track (Entity)
- disc, track, title (work+movement, NO composer)
- artists[] (includes composer, performers, conductor, arranger)
- name (filename)
- Validate() checks tags AND filename

### Artist (Value Object)
- name, role (immutable)
- Role: Composer, Soloist, Ensemble, Conductor, Arranger, Guest

### ValidationIssue (Value Object)
- level (Error/Warning/Info)
- track (0=album, -1=directory)
- rule (section number + text from rules docs)
- message (context-specific)

## Key Validation Rules
1. Mandatory tags: Composer, Artist Name, Track Title, Album Title, Track Number
2. Composer NOT in track title tag (last name check)
3. Artist format: "Soloist, Ensemble, Conductor"
4. Path length: 180 chars max
5. Filename format: `## - Track Title.flac`
6. Multi-disc: subdirectories (CD1/, CD2/)
7. Title Case required
8. Parse arranger from "(arr. by X)" in title
9. Multiple composers = error at Track construction

## Examples
- `/Pacifica Quartet - The Soviet Experience, Vol 4 (2013) [24-96 FLAC]/CD1/01 String Quartet No. 13.flac`
- `/John Eliot Gardiner - Bach_ Motets [48-24 FLAC]/01_Bach, Testament Lobet den Herrn.flac`

## Implementation Order
1. Value objects (Level, Role, Artist)
2. Entities (Edition, Track, Album)
3. Validation rules
4. Tag reader/writer
5. Directory validator
6. Web scrapers
7. CLI applications

## Decisions
- Use `github.com/dhowden/tag` for better maintenance
- Private fields with getters (immutability)
- Rules in message file
- Simple lastName() extraction
- Multiple composers fail at NewTrack()
- Separate directory validator
