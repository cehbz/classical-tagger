# Next Steps Guide

This document outlines how to proceed with the classical-tagger project.

## Immediate Actions (Do This First)

### 1. Add the Validate CLI to Your Repository

```bash
# In your repository root:
mkdir -p cmd/validate
cp path/to/outputs/cmd/validate/* cmd/validate/

# Commit the changes
git add cmd/validate/
git commit -m "feat: add validate CLI for directory validation"
git push
```

### 2. Test the Validate CLI

```bash
# Build the command
cd cmd/validate
go build -o validate

# Test with a real classical music directory
./validate "/path/to/your/classical/album"

# Example with Bach:
./validate "/music/Bach - Goldberg Variations (1981) - FLAC"
```

### 3. Run All Tests

```bash
# From repository root
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

## What Just Got Added

### New Files (3 total)
1. `cmd/validate/main.go` - Complete validation command (319 lines)
2. `cmd/validate/main_test.go` - Unit tests for scanner (85 lines)
3. `cmd/validate/README.md` - Documentation

### Key Features
- ✅ Recursive directory scanning
- ✅ Multi-disc detection (CD1, CD2, etc.)
- ✅ FLAC tag reading and validation
- ✅ Directory structure validation
- ✅ Comprehensive error reporting
- ✅ Colored output with emojis
- ✅ Exit codes for CI/CD integration

## Project Status Summary

### What's Complete ✅
1. **Domain Model** (internal/domain/) - 15 files
2. **Validation Rules** (internal/validation/) - 3 files
3. **FLAC Tag Reading** (internal/tagging/) - 2 files
4. **Filesystem Validation** (internal/filesystem/) - 2 files
5. **JSON Storage** (internal/storage/) - 2 files
6. **Validate CLI** (cmd/validate/) - 3 files ⚠️ NEW

**Total:** 27 files, ~3500 lines of production code

### What's Next 🚧

Priority order for implementation:

#### 1. Tag Writing (HIGH PRIORITY)
**Why:** Users need to apply validated metadata to files

**What to build:**
```
internal/tagging/
├── writer.go         # FLAC tag writer
├── writer_test.go    # Tests
cmd/tag/
├── main.go           # CLI for applying tags
├── main_test.go      # Tests
└── README.md         # Documentation
```

**Example usage:**
```bash
# Apply metadata from JSON to FLAC files
tag --apply metadata.json --dir /music/album

# Dry run (show what would change)
tag --dry-run metadata.json

# Backup before applying
tag --apply metadata.json --backup
```

#### 2. Web Scraping - Harmonia Mundi (HIGH PRIORITY)
**Why:** Need a data source for metadata

**What to build:**
```
internal/scraping/
├── scraper.go              # Common scraping interface
├── harmoniamund/
│   ├── extractor.go        # HM-specific parsing
│   ├── extractor_test.go   # Tests
│   └── README.md           # Documentation
cmd/extract/
├── main.go                 # CLI for extraction
├── main_test.go            # Tests
└── README.md               # Documentation
```

**Example usage:**
```bash
# Extract from single URL
extract --url https://www.harmoniamundi.com/... --output metadata.json

# Batch extraction
extract --batch urls.txt --output-dir metadata/
```

#### 3. Enhanced Parsing (MEDIUM PRIORITY)
**Why:** Improve metadata accuracy

**Improvements needed:**
- Parse "Soloist, Ensemble, Conductor" format
- Auto-detect "(arr. by X)" in titles
- Validate title capitalization
- Validate movement number formats

#### 4. Additional Scrapers (LOW PRIORITY)
**Sites to add:**
- Classical Archives
- Naxos
- Presto Classical
- ArkivMusic

## Development Workflow

### For Each New Feature

1. **Write tests first** (TDD)
```bash
# Create test file
touch internal/package/feature_test.go

# Write failing tests
# Implement minimal code to pass
# Refactor while keeping tests green
```

2. **Follow the established patterns**
- Immutable domain objects
- Private fields with getters
- Constructor functions (New*)
- Builder pattern for optional fields
- Error returns, never panic

3. **Maintain test coverage**
```bash
# Check coverage
go test -cover ./internal/package

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

4. **Update documentation**
- Add README.md to new packages
- Update IMPLEMENTATION_STATUS.md
- Add examples to package docs

## Testing With Real Data

### Create Test Fixtures

You'll need sample FLAC files for integration tests. Create this structure:

```
testdata/
├── single-disc/
│   ├── Bach - Goldberg Variations (1981) - FLAC/
│   │   ├── 01 Aria.flac
│   │   ├── 02 Variation 1.flac
│   │   └── metadata.json
├── multi-disc/
│   ├── Wagner - Der Ring des Nibelungen (1990) - FLAC/
│   │   ├── CD1/
│   │   │   ├── 01 Das Rheingold, Scene 1.flac
│   │   │   └── 02 Das Rheingold, Scene 2.flac
│   │   ├── CD2/
│   │   │   └── 01 Die Walküre, Act 1.flac
│   │   └── metadata.json
└── invalid/
    ├── missing-composer/
    ├── bad-filenames/
    └── path-too-long/
```

### Integration Test Template

```go
func TestValidateRealAlbum(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    testDir := filepath.Join("testdata", "single-disc", 
        "Bach - Goldberg Variations (1981) - FLAC")
    
    report, err := ValidateDirectory(testDir)
    if err != nil {
        t.Fatalf("Validation failed: %v", err)
    }
    
    if report.HasErrors() {
        t.Errorf("Expected no errors, got %d", 
            len(report.MetadataIssues))
    }
}
```

## Architecture Guidelines

### When Adding New Packages

1. **Follow clean architecture:**
```
internal/
├── domain/       # Business logic, entities
├── validation/   # Business rules
├── tagging/      # Infrastructure (FLAC I/O)
├── filesystem/   # Infrastructure (file system)
├── scraping/     # Infrastructure (HTTP)
└── storage/      # Infrastructure (JSON)
```

2. **Keep dependencies pointing inward:**
- cmd → internal
- infrastructure → domain
- Never: domain → infrastructure

3. **Use interfaces for external dependencies:**
```go
// Good: testable
type TagReader interface {
    ReadFile(path string) (*Metadata, error)
}

// Bad: concrete dependency
func Process(reader *FLACReader) { }
```

### Code Quality Checklist

Before committing:
- [ ] All tests pass (`go test ./...`)
- [ ] Test coverage >90% for new code
- [ ] No `TODO` comments without GitHub issues
- [ ] Documentation updated
- [ ] Examples provided where helpful
- [ ] Error messages are actionable
- [ ] Exported types have doc comments

## Common Patterns in This Codebase

### 1. Immutable Value Objects
```go
type Artist struct {
    name string
    role Role
}

// No setters, only getters
func (a Artist) Name() string { return a.name }
```

### 2. Builder Pattern
```go
track := NewTrack(disc, num, title, artists)
track = track.WithName("01 Aria.flac")
```

### 3. Validation at Construction
```go
// Fail fast
track, err := NewTrack(disc, num, title, artists)
if err != nil {
    return nil, err // Don't create invalid objects
}
```

### 4. Validation for Business Rules
```go
// Check business rules later
issues := album.Validate()
for _, issue := range issues {
    if issue.Level() == LevelError {
        // Handle error
    }
}
```

### 5. DTO Pattern
```go
// Domain model
type Album struct { ... }

// JSON representation
type AlbumDTO struct { ... }

// Conversion
func FromAlbum(album *Album) AlbumDTO { ... }
func (dto AlbumDTO) ToAlbum() (*Album, error) { ... }
```

## Debugging Tips

### Enable Verbose Output
```bash
# See all test output
go test -v ./...

# Run specific test
go test -v -run TestTrack_Validate ./internal/domain
```

### Use Table-Driven Tests
```go
tests := []struct {
    name string
    input X
    want Y
    wantErr bool
}{
    // ...
}
```

### Check Generated JSON
```go
// Pretty print JSON during tests
b, _ := json.MarshalIndent(album, "", "  ")
t.Logf("Generated JSON:\n%s", string(b))
```

## Getting Help

### Resources
1. **Project documentation:** See package-level READMEs
2. **Test examples:** Look at *_test.go files
3. **Design decisions:** See IMPLEMENTATION_STATUS.md
4. **Architecture:** See domain/ package docs

### Questions to Ask
- Is this the right layer for this code?
- Does this follow existing patterns?
- Is this testable?
- Will this be maintainable?
- Does this violate SOLID principles?

## Git Workflow

### Branch Strategy
```bash
# Feature branches
git checkout -b feat/tag-writer
git checkout -b feat/harmonia-scraper

# Bug fixes
git checkout -b fix/validation-panic

# Documentation
git checkout -b docs/update-readme
```

### Commit Messages
```bash
# Format: <type>: <description>

git commit -m "feat: add FLAC tag writer"
git commit -m "test: add integration tests for validator"
git commit -m "fix: handle missing composer tag gracefully"
git commit -m "docs: update README with new examples"
git commit -m "refactor: extract common validation logic"
```

### Before Pushing
```bash
# Run all tests
go test ./...

# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run linter if available
golangci-lint run
```

## Performance Considerations

### For Large Directories
```bash
# Profile memory usage
go test -memprofile=mem.prof ./cmd/validate

# Profile CPU usage
go test -cpuprofile=cpu.prof ./cmd/validate

# Analyze profiles
go tool pprof cpu.prof
```

### Optimization Tips
- Use buffered I/O for large files
- Parse tags lazily if possible
- Consider worker pools for parallel processing
- Cache repeated file system operations

## Security Considerations

### When Adding Web Scraping
- Validate URLs before fetching
- Set request timeouts
- Limit response sizes
- Handle redirects safely
- Respect robots.txt

### When Writing Tags
- Always validate before writing
- Create backups
- Handle partial failures
- Don't follow symlinks blindly
- Check file permissions

## Deployment

### Building Binaries
```bash
# Build all commands
go build -o bin/ ./cmd/...

# Build for specific OS
GOOS=linux GOARCH=amd64 go build -o validate-linux ./cmd/validate
GOOS=darwin GOARCH=amd64 go build -o validate-macos ./cmd/validate
GOOS=windows GOARCH=amd64 go build -o validate.exe ./cmd/validate
```

### Installation
```bash
# Install to $GOPATH/bin
go install ./cmd/validate
go install ./cmd/tag
go install ./cmd/extract
```

## Success Criteria

You'll know you're done when:
- [ ] All three CLIs work end-to-end
- [ ] Can validate a real album directory
- [ ] Can scrape metadata from Harmonia Mundi
- [ ] Can apply JSON metadata to FLAC files
- [ ] All tests pass
- [ ] Test coverage >95%
- [ ] Documentation is complete
- [ ] No critical bugs
- [ ] Performance is acceptable (<1s per album)

## Final Checklist Before v1.0

- [ ] All planned features implemented
- [ ] Integration tests with real data
- [ ] Performance testing
- [ ] User documentation
- [ ] Installation guide
- [ ] Example workflows
- [ ] Error handling review
- [ ] Security review
- [ ] Code review
- [ ] Beta testing with real users

---

Good luck! The foundation is solid - now it's time to build the rest of the features on top of it.
