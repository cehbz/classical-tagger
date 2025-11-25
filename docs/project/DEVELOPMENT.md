# Development Guide

This document describes coding standards, development practices, and contribution guidelines for the classical-tagger project.

## Table of Contents

- [Getting Started](#getting-started)
- [Coding Standards](#coding-standards)
- [Test-Driven Development](#test-driven-development)
- [Domain-Driven Design](#domain-driven-design)
- [Git Workflow](#git-workflow)
- [Code Review](#code-review)
- [Adding Features](#adding-features)

## Getting Started

### Prerequisites

```bash
# Install Go 1.25+
go version  # Should show go1.25 or higher

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install mktorrent (for upload functionality)
# Ubuntu/Debian:
sudo apt-get install mktorrent
# macOS:
brew install mktorrent
```

### Clone and Build

```bash
git clone https://github.com/cehbz/classical-tagger
cd classical-tagger

# Run tests to verify setup
go test ./...

# Build all commands
go build ./cmd/validate
go build ./cmd/extract
go build ./cmd/tag
go build ./cmd/upload
```

### IDE Setup

**VS Code / Cursor:**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

## Coding Standards

### Go Code Style

**Follow standard Go conventions:**
- Use `gofmt` (or `goimports`) for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

**Key rules:**

1. **Naming:**
   - Exported: `PascalCase`
   - Unexported: `camelCase`
   - Acronyms: `APIKey`, `HTTPClient` (not `ApiKey`, `HttpClient`)
   - Interfaces: `-er` suffix when possible (`Validator`, `Extractor`)

2. **Comments:**
   - Every exported symbol has a doc comment
   - Doc comments start with the symbol name
   - Full sentences with proper punctuation
   ```go
   // Validator runs validation rules against torrents.
   type Validator struct { ... }
   
   // Validate checks a torrent against all registered rules.
   func (v *Validator) Validate(t *domain.Torrent) []domain.ValidationIssue {
   ```

3. **Error Handling:**
   - Check all errors, never ignore
   - Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
   - Return early on errors (avoid deep nesting)
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to load config: %w", err)
   }
   
   // Bad
   if err == nil {
       // ... lots of nested code
   }
   ```

4. **Package Organization:**
   - One concept per package
   - Avoid circular dependencies
   - Keep packages focused and cohesive

### Project-Specific Conventions

1. **Domain Objects:**
   - Use exported fields (not getters)
   - Value objects are immutable
   - No business logic in domain structs
   ```go
   // Good - exported fields
   type Track struct {
       Disc   int
       Track  int
       Title  string
   }
   
   // Bad - getters (old style)
   type Track struct {
       disc  int
       track int
   }
   func (t *Track) Disc() int { return t.disc }
   ```

2. **Validation Rules:**
   - Method naming: `Rule_RuleName`
   - Return empty slice if valid
   - Return all issues, not just first
   ```go
   func (v *Validator) Rule_ComposerNotInTitle(t *domain.Torrent) []domain.ValidationIssue {
       var issues []domain.ValidationIssue
       // ... check logic
       return issues
   }
   ```

3. **Testing:**
   - Table-driven tests for multiple cases
   - Descriptive test names: `TestValidator_Rule_ComposerNotInTitle`
   - Test both success and failure cases
   - Use `testdata/` for fixtures

4. **Configuration:**
   - Single source: `~/.config/classical-tagger/config.yaml`
   - No environment variables (except XDG standard)
   - Command-line flags override config

## Test-Driven Development

### TDD Workflow

**Red-Green-Refactor Cycle:**

1. **Red:** Write a failing test
   ```go
   func TestTrack_HasComposer(t *testing.T) {
       track := &domain.Track{
           Artists: []domain.Artist{
               {Name: "Bach", Role: domain.RoleComposer},
           },
       }
       
       if !track.HasComposer() {
           t.Error("Expected track to have composer")
       }
   }
   ```

2. **Green:** Write minimal code to pass
   ```go
   func (t *Track) HasComposer() bool {
       for _, artist := range t.Artists {
           if artist.Role == RoleComposer {
               return true
           }
       }
       return false
   }
   ```

3. **Refactor:** Improve code while keeping tests green
   ```go
   // Maybe extract to a helper function if used elsewhere
   func hasRole(artists []Artist, role Role) bool {
       for _, artist := range artists {
           if artist.Role == role {
               return true
           }
       }
       return false
   }
   ```

### Testing Patterns

**Table-Driven Tests:**
```go
func TestSanitizeFilename(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {
            name:  "remove invalid characters",
            input: "Track: \"Name\" / Path",
            want:  "Track Name Path",
        },
        {
            name:  "trim spaces",
            input: "  Title  ",
            want:  "Title",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := SanitizeFilename(tt.input)
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

**Test Helpers:**
```go
// Create test fixtures
func newTestTorrent() *domain.Torrent {
    return &domain.Torrent{
        Title:        "Test Album",
        OriginalYear: 2020,
        Files: []domain.FileLike{
            &domain.Track{
                File:  domain.File{Path: "01 - Track.flac"},
                Disc:  1,
                Track: 1,
                Title: "Track Title",
            },
        },
    }
}

// Use in tests
func TestSomething(t *testing.T) {
    torrent := newTestTorrent()
    // ... test code
}
```

**Mocking External Dependencies:**
```go
// Interface for mocking
type HTTPClient interface {
    Get(url string) (*http.Response, error)
}

// Mock implementation
type mockHTTPClient struct {
    response *http.Response
    err      error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
    return m.response, m.err
}
```

### Test Coverage

**Goals:**
- Domain logic: 100%
- Validation rules: 100%
- Extractors: 80%+
- CLIs: Integration tests

**Check coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Domain-Driven Design

### Ubiquitous Language

**Use domain terminology consistently:**

| Term | Meaning | Usage |
|------|---------|-------|
| Torrent | Collection of files | Aggregate root |
| Track | Audio file with metadata | Entity within torrent |
| Work | Musical composition | Album or multi-movement piece |
| Movement | Part of a work | Individual track |
| Composer | Creator of music | Artist role |
| Performer | Soloist, ensemble, conductor | Artist roles |
| Edition | Specific release | Label, catalog number, year |
| Trump | Replace existing torrent | Upload command context |

### Domain Layer Purity

**The domain package has no dependencies:**
```go
// ✅ Good - pure domain
package domain

type Torrent struct {
    Title        string
    OriginalYear int
    Files        []FileLike
}

// ❌ Bad - domain depends on infrastructure
package domain

import "database/sql"

type Torrent struct {
    Title string
    db    *sql.DB  // NO! Domain should not know about DB
}
```

### Aggregates

**Rules for working with aggregates:**

1. External code gets aggregate root references only
2. Modifications go through aggregate root methods
3. Aggregates enforce their own invariants

```go
// ✅ Good - access through aggregate
torrent := loadTorrent()
tracks := torrent.Tracks()  // Returns []Track

// ❌ Bad - direct entity access
tracks := loadTracksDirectly()  // Bypasses aggregate
```

## Git Workflow

### Branch Strategy

```bash
# Create feature branch
git checkout -b feature/add-naxos-extractor

# Make changes, commit frequently
git add .
git commit -m "Add Naxos extractor interface"

# Keep up to date with main
git fetch origin
git rebase origin/main

# Push when ready
git push -u origin feature/add-naxos-extractor
```

### Commit Messages

**Format:**
```
Short summary (50 chars or less)

Detailed explanation if needed:
- What changed
- Why it changed
- Any side effects

Fixes #123
```

**Examples:**
```
Add Naxos metadata extractor

Implements scraper for naxos.com following the pattern
established for Discogs. Includes comprehensive tests
and integration with the extract CLI.

Fixes #45
```

### Pull Requests

**Checklist before creating PR:**
- [ ] All tests pass: `go test ./...`
- [ ] Linting passes: `golangci-lint run`
- [ ] Code is formatted: `gofmt -s -w .`
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (if user-facing)

## Code Review

### Reviewing Code

**What to look for:**

1. **Correctness:**
   - Does it solve the problem?
   - Are edge cases handled?
   - Are errors handled properly?

2. **Tests:**
   - Are there tests?
   - Do tests cover the important cases?
   - Are tests clear and maintainable?

3. **Design:**
   - Is it consistent with existing patterns?
   - Does it follow SOLID principles?
   - Is the API intuitive?

4. **Style:**
   - Is it formatted properly?
   - Are names clear?
   - Are comments helpful?

### Responding to Reviews

**Be receptive:**
- Reviews make code better
- Ask questions if you don't understand feedback
- Explain your reasoning if you disagree
- Don't take it personally

## Adding Features

### Adding a Validation Rule

1. Add test in `internal/validation/validator_test.go`:
   ```go
   func TestValidator_Rule_NewRule(t *testing.T) {
       v := NewValidator()
       torrent := &domain.Torrent{...}
       
       issues := v.Rule_NewRule(torrent)
       // Assert expected issues
   }
   ```

2. Implement rule in `internal/validation/validator.go`:
   ```go
   func (v *Validator) Rule_NewRule(t *domain.Torrent) []domain.ValidationIssue {
       var issues []domain.ValidationIssue
       // Check logic
       return issues
   }
   ```

3. Rule is automatically discovered and run

### Adding a Metadata Extractor

See [Adding Scrapers Guide](docs/development/adding-scrapers.md) for detailed steps.

**Quick overview:**
1. Create `internal/scraping/newsite_extractor.go`
2. Implement `Extractor` interface
3. Add tests
4. Register in CLI

### Adding a CLI Command

1. Create directory: `cmd/newcommand/`
2. Create `main.go` with flag parsing
3. Call application service from `internal/`
4. Add tests
5. Update documentation

## Troubleshooting

### Common Issues

**Import cycle:**
- Move shared code to a new package
- Check dependency direction (domain → nothing)

**Test failures:**
- Run with `-v` for verbose output
- Use `t.Logf()` for debugging
- Check test isolation (are tests affecting each other?)

**Linter errors:**
- Run `golangci-lint run` to see all issues
- Fix or add `//nolint:rulename` with justification

### Getting Help

- Check existing issues on GitHub
- Ask in GitHub Discussions
- Read the architecture documentation
- Look at similar code in the project

## Resources

### Go Language
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

### Design Patterns
- Domain-Driven Design (Eric Evans)
- Clean Architecture (Robert C. Martin)
- Test-Driven Development (Kent Beck)

### Project-Specific
- [Architecture](ARCHITECTURE.md)
- [TODO](TODO.md)
- [Metadata Sources](../development/metadata-sources.md)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.