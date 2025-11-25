# Architecture

This document describes the domain model, design decisions, and architectural patterns used in the classical-tagger project.

## Table of Contents

- [Domain Model](#domain-model)
- [Aggregate Boundaries](#aggregate-boundaries)
- [Design Principles](#design-principles)
- [Package Structure](#package-structure)
- [Key Architectural Decisions](#key-architectural-decisions)

## Domain Model

### Core Concept: Torrent as Aggregate Root

The domain model centers on **Torrent** as the aggregate root, not "Album". This is a critical insight:

**Why Torrent, not Album?**
- A torrent is a collection of **files** representing an album
- Contains both **audio files** (tracks with metadata) and **non-audio files** (logs, cues, artwork)
- Leverages Go's type system: `Track` embeds `File` rather than maintaining separate collections
- Better reflects the actual domain: we're managing torrents, not abstract albums

### Domain Entities

#### Torrent (Aggregate Root)
```go
type Torrent struct {
    RootPath     string      // Location on disk
    Title        string      // Album title
    OriginalYear int         // Original release year
    Edition      *Edition    // Release edition info
    AlbumArtist  []Artist    // Album-level artists
    Files        []FileLike  // All files (mix of File and Track)
    SiteMetadata *SiteMetadata // Redacted-specific metadata
}
```

**Responsibilities:**
- Maintains consistency of album-level metadata
- Contains all files in the torrent
- Provides methods to access tracks specifically

**Invariants:**
- Title must not be empty
- OriginalYear must be positive
- Files list contains both tracks and non-audio files

#### Track (Entity, embeds File)
```go
type Track struct {
    File             // Embedded: provides Path field
    Disc     int     // Disc number
    Track    int     // Track number
    Title    string  // Track title
    Artists  []Artist // Track-level artists
}
```

**Responsibilities:**
- Represents a single audio file with metadata
- Embeds File to inherit path information
- Contains artist credits specific to this track

**Invariants:**
- Disc and Track numbers must be positive
- Title must not be empty
- Must have at least one artist (typically composer)

#### File (Value Object)
```go
type File struct {
    Path string // Relative path from torrent root
}
```

**Responsibilities:**
- Represents a non-audio file (log, cue, artwork, etc.)
- Provides path information

**Invariants:**
- Path must not be empty

### Value Objects

#### Artist
```go
type Artist struct {
    Name string
    Role Role // Composer, Soloist, Ensemble, Conductor, etc.
}
```

**Immutability:** Artists are immutable value objects. Two artists with same name and role are considered equal.

**Roles:**
```go
type Role int

const (
    RoleComposer Role = iota
    RoleSoloist
    RoleEnsemble
    RoleConductor
    RoleArranger
    RoleGuest
)
```

#### Edition
```go
type Edition struct {
    Label         string
    CatalogNumber string
    Year          int
}
```

**Purpose:** Represents a specific release edition, distinct from the original work.

#### ValidationIssue
```go
type ValidationIssue struct {
    Level   Level  // ERROR, WARNING, INFO
    Track   int    // Track number (0 for album-level)
    Rule    string // Rule name that failed
    Message string // Human-readable message
}
```

**Immutability:** Validation issues are immutable value objects.

## Aggregate Boundaries

### Torrent Aggregate

**Root:** `Torrent`

**Entities within aggregate:**
- `Track` (multiple)

**Value objects within aggregate:**
- `File` (multiple)
- `Artist` (multiple, on album and track level)
- `Edition` (optional)
- `SiteMetadata` (optional)

**Boundary rules:**
- External code gets `Torrent` references, never direct `Track` references
- Modifications to tracks go through `Torrent` methods
- Tracks cannot exist without their parent `Torrent`

### Why this boundary?

- **Consistency:** Album-level metadata (title, year) must be consistent with all tracks
- **Completeness:** A torrent is complete only when all files are present
- **Validation:** Validation rules often span multiple tracks (e.g., disc numbering)

## Design Principles

### 1. Domain-Driven Design (DDD)

**Ubiquitous Language:**
- Terms match the classical music domain: "work", "movement", "opus", "catalog number"
- Terms match the torrent domain: "torrent", "tracker", "trump"

**Bounded Contexts:**
- **Core Domain:** Torrent, Track, Artist (in `internal/domain`)
- **Validation Context:** Rules, issues, levels (in `internal/validation`)
- **Extraction Context:** Scrapers, parsers (in `internal/scraping`)
- **Persistence Context:** Repository, serialization (in `internal/storage`)

**Domain Services:**
- Validation engine (`internal/validation`)
- Tag reader/writer (`internal/tagging`)
- Metadata extractors (`internal/scraping`)

### 2. Test-Driven Development (TDD)

**Red-Green-Refactor:**
1. Write failing test
2. Implement minimal code to pass
3. Refactor while keeping tests green

**Test Coverage Goals:**
- Domain logic: 100%
- Validation rules: 100%
- Extractors: 80%+ (some HTML parsing is hard to mock)
- CLIs: Integration tests

### 3. SOLID Principles

**Single Responsibility:**
- `Torrent`: Maintains album data
- `Validator`: Runs validation rules
- `TagReader`: Reads FLAC tags
- `Repository`: Persists to JSON

**Open/Closed:**
- Validation rules are discovered via reflection, new rules added without modifying engine
- Extractors implement common interface, new sources added without modifying framework

**Liskov Substitution:**
- `FileLike` interface allows `File` and `Track` to be used interchangeably in collections
- All extractors implement `Extractor` interface

**Interface Segregation:**
- `TagReader` separate from `TagWriter`
- `Validator` separate from `ValidationRule`

**Dependency Inversion:**
- Domain depends on nothing
- Application services depend on domain
- Infrastructure (CLI, HTTP) depends on application services

### 4. Immutability (Evolved Approach)

**Original Approach:** Immutable objects with getter methods.

**Current Approach:** Mutable objects with exported fields.

**Why the change?**
- Go is not a functional language; immutability adds unnecessary complexity
- Getter methods create boilerplate without safety benefits
- Exported fields with careful API design provides better ergonomics
- Tests and domain logic are clearer with direct field access

**Where immutability remains:**
- Value objects (`Artist`, `ValidationIssue`)
- Configuration objects
- Extraction results (return new objects, don't mutate inputs)

## Package Structure

### `/internal/domain`
**Purpose:** Core domain model, no external dependencies.

**Contains:**
- `Torrent` (aggregate root)
- `Track`, `File` (entities)
- `Artist`, `Edition`, `ValidationIssue` (value objects)
- `Role`, `Level` (enums)

**Rules:**
- Pure data structures
- No business logic (validation, persistence, etc.)
- No dependencies on other internal packages

### `/internal/validation`
**Purpose:** Validation rules engine.

**Contains:**
- `Validator` - Discovers and runs rules
- `ValidatorTest` - Test utilities
- Rule methods (discovered via reflection)

**Key Pattern:** Reflection-based rule discovery inspired by Go's testing framework.

```go
// Rules are methods matching signature:
func (v *Validator) Rule_RuleName(t *domain.Torrent) []domain.ValidationIssue
```

### `/internal/tagging`
**Purpose:** FLAC tag reading and writing.

**Contains:**
- `TagReader` - Reads tags from FLAC files
- `TagWriter` - Writes tags to FLAC files
- Filename utilities

**Dependencies:**
- `go-flac/go-flac` for FLAC file handling
- `go-flac/flacvorbis` for Vorbis comments

### `/internal/scraping`
**Purpose:** Web metadata extraction.

**Contains:**
- `Extractor` interface
- Site-specific extractors (Discogs, Harmonia Mundi, etc.)
- `LocalExtractor` - Extracts from existing FLAC files

**Pattern:** Each extractor is independent, registered in a central registry.

### `/internal/storage`
**Purpose:** Persistence (JSON serialization).

**Contains:**
- `Repository` - Serializes/deserializes `Torrent`

**Design:** No DTOs—domain objects serialize directly with JSON tags.

### `/internal/config`
**Purpose:** Configuration management.

**Contains:**
- Config loading (XDG-compliant)
- API token management

**Rules:**
- Single source of truth: `~/.config/classical-tagger/config.yaml`
- No environment variables (except XDG standard paths)
- Command-line overrides for operational parameters only

### `/internal/uploader`
**Purpose:** Redacted upload logic.

**Contains:**
- `UploadCommand` - Orchestrates upload workflow
- API client for Redacted
- Metadata merging logic

### `/cmd/*`
**Purpose:** Command-line applications.

**Pattern:** Each CLI is a thin layer:
1. Parse flags
2. Load configuration
3. Call application service
4. Format output

## Key Architectural Decisions

### ADR-001: Torrent as Aggregate Root

**Context:** Original design used "Album" as the aggregate root.

**Decision:** Changed to "Torrent" as the aggregate root.

**Rationale:**
- Torrents contain files (both audio and non-audio)
- Better reflects the actual domain
- Leverages Go's type embedding (`Track` embeds `File`)
- Simpler than maintaining separate file and track collections

**Consequences:**
- Domain model better matches reality
- Code is more intuitive
- Easier to reason about file operations

### ADR-002: Mutable Objects with Exported Fields

**Context:** Original design used immutable objects with getter methods.

**Decision:** Use mutable objects with exported fields.

**Rationale:**
- Go is not a functional language
- Immutability adds complexity without benefits in this domain
- Getter methods create boilerplate
- Tests and domain logic are clearer with direct access

**Consequences:**
- More idiomatic Go code
- Better ergonomics for developers
- Requires discipline to not abuse mutability

### ADR-003: Reflection-Based Validation Rules

**Context:** Need extensible validation system.

**Decision:** Use reflection to discover validation rule methods.

**Rationale:**
- Similar to Go's testing framework (familiar pattern)
- No manual rule registration
- Type-safe (compile-time method signature checking)
- Easy to add new rules

**Consequences:**
- Rules are discovered automatically
- Slight runtime overhead (minimal, one-time on startup)
- Requires naming convention (`Rule_*` prefix)

### ADR-004: No DTOs for Persistence

**Context:** Need to serialize domain objects to JSON.

**Decision:** Use JSON tags on domain objects directly.

**Rationale:**
- Domain objects are simple data structures
- No impedance mismatch between domain and persistence
- Less code, less maintenance
- JSON format is stable and human-readable

**Consequences:**
- Domain objects have JSON tags (minor coupling)
- Changes to domain structure may affect JSON format
- Acceptable tradeoff for simplicity

### ADR-005: Single XDG Config File

**Context:** Need to manage API tokens and configuration.

**Decision:** Single config file at `~/.config/classical-tagger/config.yaml`, no environment variables.

**Rationale:**
- Single source of truth
- XDG-compliant (standard Linux/Unix convention)
- Easy to manage and version control (user choice)
- Command-line overrides for operational parameters

**Consequences:**
- Users must create config file
- No "magic" environment variables
- Clear configuration location
- Tool provides helpful error messages pointing to config file location

### ADR-006: Rate Limiting with "OnResponse" Semantics

**Context:** API integrations (Discogs, Redacted) have rate limits.

**Decision:** Rate limiter with "OnResponse" semantics (sliding window based on actual request-response cycles).

**Rationale:**
- Simple request throttling can overwhelm slow servers
- Sliding window based on response prevents server overload
- More respectful to APIs

**Consequences:**
- Slightly more complex rate limiter implementation
- Better behavior with real-world APIs
- Prevents rate limit violations

## Testing Strategy

### Unit Tests
- Domain logic (100% coverage goal)
- Validation rules (100% coverage goal)
- Value object behavior

### Integration Tests
- Tag reading/writing with real FLAC files
- JSON serialization/deserialization
- CLI argument parsing

### System Tests
- End-to-end workflows (extract → validate → tag → upload)
- Real FLAC files in `testdata/`
- Skippable network tests for web scraping

## Future Considerations

### Potential Improvements

1. **Event Sourcing** (if needed):
   - Track all changes to torrents
   - Enable undo/redo
   - Audit trail for uploads

2. **CQRS** (if needed):
   - Separate read and write models
   - Optimize queries (e.g., search)
   - Useful if adding web UI

3. **Plugin Architecture**:
   - Dynamic loading of validation rules
   - User-defined extractors
   - Custom formatters

### Constraints

- Go's type system (no generics in older versions, limited in 1.18+)
- FLAC library limitations
- API rate limits
- XDG paths may not work on Windows (need fallback)

## References

- Domain-Driven Design: Eric Evans
- Clean Architecture: Robert C. Martin
- Go Proverbs: Rob Pike
- Redacted Classical Music Upload Guide
