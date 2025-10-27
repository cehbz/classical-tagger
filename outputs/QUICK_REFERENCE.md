# Quick Reference Card

## Common Commands

### Testing
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...

# Specific package
go test -v ./internal/domain

# Run specific test
go test -v -run TestTrack_Validate ./internal/domain

# With race detection
go test -race ./...

# Short mode (skip slow tests)
go test -short ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building
```bash
# Build validate CLI
cd cmd/validate && go build -o validate

# Build all CLIs
go build -o bin/ ./cmd/...

# Install to $GOPATH/bin
go install ./cmd/validate
```

### Code Quality
```bash
# Format code
go fmt ./...

# Check for issues
go vet ./...

# Run linter (if installed)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Quick Architecture Reference

```
Domain Model (internal/domain/)
├── Level         → ERROR, WARNING, INFO
├── Role          → Composer, Soloist, Ensemble, Conductor, Arranger
├── Artist        → Immutable (name, role)
├── ValidationIssue → (level, track, rule, message)
├── Edition       → (label, catalogNumber, year)
├── Track         → Entity (disc, track, title, artists[], name)
└── Album         → Aggregate root (title, year, edition, tracks[])

Infrastructure
├── validation/   → AlbumValidator, rules database
├── tagging/      → FLACReader (writer TODO)
├── filesystem/   → DirectoryValidator
├── storage/      → JSON serialization (DTO pattern)
└── scraping/     → (TODO)

Commands
├── validate/     → ✅ Directory validator
├── tag/          → 🚧 Tag applicator
└── extract/      → 🚧 Web scraper
```

## Common Patterns

### Creating Entities
```go
// Track
track, err := domain.NewTrack(1, 1, "Aria", []domain.Artist{composer})
if err != nil {
    return err // Fail fast
}
track = track.WithName("01 Aria.flac") // Builder pattern

// Album
album, err := domain.NewAlbum("Goldberg Variations", 1981)
album = album.AddTrack(track)
```

### Validation
```go
// At construction (fail fast)
track, err := domain.NewTrack(disc, num, title, artists)
// err != nil means invalid state impossible

// Business rules (later)
issues := album.Validate()
for _, issue := range issues {
    switch issue.Level {
    case domain.LevelError:
        // Critical
    case domain.LevelWarning:
        // Recommended
    case domain.LevelInfo:
        // Suggestion
    }
}
```

### JSON Conversion
```go
// Domain → JSON
repo := storage.NewRepository()
jsonBytes, err := repo.SaveToJSON(album)

// JSON → Domain
album, err := repo.LoadFromJSON(jsonBytes)
```

### Reading Tags
```go
reader := tagging.NewFLACReader()

// Read metadata
metadata, err := reader.ReadFile("01 Aria.flac")

// Read and convert to Track
track, err := reader.ReadTrackFromFile("01 Aria.flac", 1, 1)
```

## Test Template

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   X
        want    Y
        wantErr bool
    }{
        {
            Name:    "valid case",
            input:   validInput,
            want:    expectedOutput,
            WantErr: false,
        },
        {
            Name:    "invalid case",
            input:   invalidInput,
            want:    nil,
            WantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.Name, func(t *testing.T) {
            got, err := Function(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", 
                    err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Function() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Validation Issue Levels

```go
// Construction errors (fail immediately)
track, err := domain.NewTrack(1, 1, "", artists) // Error: empty title

// Business rule errors (report later)
domain.LevelError   // ❌ Must fix (blocks upload)
domain.LevelWarning // ⚠️  Should fix (recommended)
domain.LevelInfo    // ℹ️  Nice to fix (suggestion)
```

## File Locations

```
Production Code:
internal/domain/            # Core business logic
internal/validation/        # Validation rules
internal/tagging/          # FLAC I/O
internal/filesystem/       # Directory validation
internal/storage/          # JSON persistence
cmd/validate/              # Validate CLI

Tests:
*_test.go files alongside code

Documentation:
README.md                  # Project overview
PROJECT_SUMMARY.md         # Current status
IMPLEMENTATION_STATUS.md   # Detailed progress
NEXT_STEPS.md             # How to continue
cmd/validate/README.md    # Validate docs
```

## Critical Rules Quick Reference

```
2.3.12     Path length ≤ 180 chars
2.3.16.4   Required tags: Composer, Artist, Album, Title, Track#
2.3.20     No leading spaces in paths
classical.composer   Composer NOT in track title
classical.folder     Composer mentioned in folder name
```

## Common Issues & Solutions

### "Multiple composers"
```go
// Wrong
artists := []domain.Artist{composer1, composer2, ...}

// Right (for now)
// Only ONE composer per track
composer := mainComposer
// Others go in title: "Work by X (arr. by Y)"
```

### "Path too long"
```
ERROR [Directory] [2.3.12] Path exceeds 180 characters

Solution: Shorten folder name or album path
```

### "Composer in title"
```
ERROR [Track 1] [classical.composer] Composer surname found in title

Wrong: "Bach: Goldberg Variations"
Right: "Goldberg Variations" (composer in Composer tag)
```

## Next Task Priorities

1. **HIGH:** Implement tag writer (internal/tagging/writer.go)
2. **HIGH:** Implement tag CLI (cmd/tag/)
3. **HIGH:** Harmonia Mundi scraper (internal/scraping/harmoniamund/)
4. **MEDIUM:** Artist parsing enhancement
5. **MEDIUM:** Arranger detection
6. **LOW:** Additional scrapers

## Useful Go Commands

```bash
# Update dependencies
go get -u ./...

# Vendor dependencies
go mod vendor

# Check for available updates
go list -u -m all

# Remove unused dependencies
go mod tidy

# View module info
go mod graph

# Download dependencies
go mod download
```

## Git Workflow

```bash
# Feature branch
git checkout -b feat/tag-writer

# Commit
git add .
git commit -m "feat: add FLAC tag writer"

# Push
git push origin feat/tag-writer

# Common commit types:
feat:     New feature
fix:      Bug fix
test:     Add tests
docs:     Documentation
refactor: Code improvement
style:    Formatting
chore:    Maintenance
```

## Performance Debugging

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Benchmarks
go test -bench=. ./...
go test -bench=. -benchmem ./...
```

## Emergency Commands

```bash
# Clean build cache
go clean -cache

# Clean test cache
go clean -testcache

# Clean module cache
go clean -modcache

# Rebuild everything
go clean && go build ./...

# Fix import issues
go mod tidy && go get ./...
```

## Quick File Creation

```bash
# New package
mkdir internal/newpkg
touch internal/newpkg/{newpkg.go,newpkg_test.go}

# New command
mkdir cmd/newcmd
touch cmd/newcmd/{main.go,main_test.go,README.md}
```

## SOLID Principles Applied

```
S - Single Responsibility
    Each validator handles ONE concern
    
O - Open/Closed
    Rules database extensible without modification
    
L - Liskov Substitution
    All validators implement same interface
    
I - Interface Segregation
    Small, focused interfaces
    
D - Dependency Inversion
    Domain has no infrastructure dependencies
```

## Remember

- ✅ **Test first** (TDD)
- ✅ **Fail fast** (construction errors)
- ✅ **Validate later** (business rules)
- ✅ **Immutable** (no setters)
- ✅ **Builder pattern** (optional fields)
- ✅ **Return errors** (never panic)
- ✅ **Document** (godoc comments)
- ✅ **Keep it simple** (no premature optimization)
