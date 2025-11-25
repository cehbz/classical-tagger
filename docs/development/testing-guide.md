# Testing Guide

This guide covers testing patterns, best practices, and examples for the classical-tagger project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Organization](#test-organization)
- [Unit Testing](#unit-testing)
- [Integration Testing](#integration-testing)
- [Test Fixtures](#test-fixtures)
- [Mocking](#mocking)
- [Coverage](#coverage)
- [Common Patterns](#common-patterns)

---

## Testing Philosophy

### Test-Driven Development (TDD)

We follow strict TDD:
1. **Red:** Write failing test first
2. **Green:** Write minimal code to pass
3. **Refactor:** Improve code while keeping tests green

### Testing Principles

1. **Tests are documentation** - They show how code should be used
2. **Tests are safety nets** - They catch regressions
3. **Tests are specifications** - They define expected behavior
4. **Fast tests** - Unit tests should run in milliseconds
5. **Isolated tests** - No dependencies between tests
6. **Deterministic tests** - Same input, same output, every time

### Test Coverage Goals

- **Domain logic:** 100%
- **Validation rules:** 100%
- **Extractors:** 80%+
- **Tag readers/writers:** 80%+
- **CLIs:** Integration tests

---

## Test Organization

### File Naming

```
internal/domain/
├── torrent.go
├── torrent_test.go        # Tests for torrent.go
├── track.go
└── track_test.go          # Tests for track.go
```

**Convention:** `filename_test.go` for `filename.go`

### Package Naming

```go
// Same package for accessing private members
package domain

func TestTorrent_Tracks(t *testing.T) {
	// Can access private fields if needed
}
```

```go
// External package for black-box testing
package domain_test

import "github.com/cehbz/classical-tagger/internal/domain"

func TestTorrent_PublicAPI(t *testing.T) {
	// Only access public API
}
```

**Guideline:** Use same package for unit tests, external package for integration tests.

### Test Function Naming

```go
// Pattern: Test<Type>_<Method>_<Scenario>
func TestValidator_Rule_ComposerNotInTitle(t *testing.T) { }
func TestValidator_Rule_ComposerNotInTitle_WithMultipleComposers(t *testing.T) { }
func TestTorrent_Tracks_EmptyFiles(t *testing.T) { }
```

---

## Unit Testing

### Table-Driven Tests

**Pattern:**
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
		{
			name:  "handle empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
```

**Benefits:**
- Easy to add new test cases
- Clear test names
- Isolated failures
- Self-documenting

### Testing Value Objects

```go
func TestArtist_Equals(t *testing.T) {
	a1 := domain.Artist{Name: "Bach", Role: domain.RoleComposer}
	a2 := domain.Artist{Name: "Bach", Role: domain.RoleComposer}
	a3 := domain.Artist{Name: "Mozart", Role: domain.RoleComposer}

	if a1 != a2 {
		t.Error("Equal artists should be equal")
	}

	if a1 == a3 {
		t.Error("Different artists should not be equal")
	}
}
```

### Testing Domain Logic

```go
func TestTorrent_Tracks(t *testing.T) {
	torrent := &domain.Torrent{
		Files: []domain.FileLike{
			&domain.Track{
				File:  domain.File{Path: "01.flac"},
				Track: 1,
			},
			&domain.File{Path: "cover.jpg"}, // Non-audio file
			&domain.Track{
				File:  domain.File{Path: "02.flac"},
				Track: 2,
			},
		},
	}

	tracks := torrent.Tracks()

	if len(tracks) != 2 {
		t.Errorf("Tracks() returned %d tracks, want 2", len(tracks))
	}

	if tracks[0].Track != 1 {
		t.Errorf("First track number = %d, want 1", tracks[0].Track)
	}
}
```

### Testing Error Conditions

```go
func TestLoadConfig_MissingFile(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	_, err := config.LoadDiscogsToken()
	if err == nil {
		t.Error("Expected error for missing config file")
	}

	// Check error message contains helpful info
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("Error message should mention config file: %v", err)
	}
}
```

---

## Integration Testing

### Testing with Real FLAC Files

```go
func TestTagReader_ReadFile(t *testing.T) {
	// Use test fixture
	reader := NewTagReader()
	tags, err := reader.ReadFile("testdata/sample.flac")
	
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if tags.Title == "" {
		t.Error("Title should not be empty")
	}
}
```

### Testing with Temporary Files

```go
func TestTagWriter_WriteFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir() // Automatically cleaned up

	// Copy test file
	src := "testdata/sample.flac"
	dst := filepath.Join(tmpDir, "output.flac")
	copyFile(t, src, dst)

	// Test writing
	writer := NewTagWriter()
	err := writer.WriteFile(dst, Tags{Title: "New Title"})
	
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify
	reader := NewTagReader()
	tags, err := reader.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	
	if tags.Title != "New Title" {
		t.Errorf("Title = %q, want %q", tags.Title, "New Title")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	
	input, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source: %v", err)
	}
	
	if err := os.WriteFile(dst, input, 0644); err != nil {
		t.Fatalf("Failed to write destination: %v", err)
	}
}
```

### Network-Dependent Tests

```go
func TestExtractor_Extract_Integration(t *testing.T) {
	// Skip in short mode (CI, quick tests)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	extractor := NewDiscogsExtractor()
	result, err := extractor.Extract("https://www.discogs.com/release/1234567")

	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result.Torrent.Title == "" {
		t.Error("Title should not be empty")
	}

	t.Logf("Extracted: %s (%d tracks)", 
		result.Torrent.Title, 
		len(result.Torrent.Files))
}
```

**Run tests:**
```bash
# Skip integration tests
go test -short ./...

# Run all tests including integration
go test ./...
```

---

## Test Fixtures

### Directory Structure

```
testdata/
├── sample.flac              # Single FLAC file
├── album/                   # Complete album
│   ├── 01 - Track.flac
│   ├── 02 - Track.flac
│   └── cover.jpg
├── multi-disc/              # Multi-disc album
│   ├── CD1/
│   │   ├── 01 - Track.flac
│   │   └── 02 - Track.flac
│   └── CD2/
│       ├── 01 - Track.flac
│       └── 02 - Track.flac
└── metadata/                # JSON fixtures
    ├── valid.json
    └── invalid.json
```

### Creating Test FLAC Files

```go
// Helper to create minimal valid FLAC file for testing
func createTestFLAC(t *testing.T, path string) {
	t.Helper()
	
	// Minimal FLAC header
	header := []byte{
		0x66, 0x4C, 0x61, 0x43, // "fLaC"
		0x80, 0x00, 0x00, 0x22, // STREAMINFO block
		// ... minimal FLAC data
	}
	
	if err := os.WriteFile(path, header, 0644); err != nil {
		t.Fatalf("Failed to create test FLAC: %v", err)
	}
}
```

**Note:** For real testing, use actual FLAC files in `testdata/`. These are NOT committed to git due to size, but documented.

### JSON Fixtures

```go
func TestRepository_LoadFromFile(t *testing.T) {
	repo := NewRepository()
	
	torrent, err := repo.LoadFromFile("testdata/metadata/valid.json")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if torrent.Title != "Expected Title" {
		t.Errorf("Title = %q, want %q", torrent.Title, "Expected Title")
	}
}
```

### Creating Fixtures

**testdata/metadata/valid.json:**
```json
{
  "root_path": "test-album",
  "title": "Test Album",
  "original_year": 2020,
  "files": [
    {
      "path": "01 - Track.flac",
      "disc": 1,
      "track": 1,
      "title": "Track Title",
      "artists": [
        {
          "name": "Composer",
          "role": 0
        }
      ]
    }
  ]
}
```

---

## Mocking

### Mock HTTP Client

```go
type mockHTTPClient struct {
	responses map[string]*http.Response
	err       error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	resp, ok := m.responses[url]
	if !ok {
		return nil, fmt.Errorf("no mock response for %s", url)
	}
	
	return resp, nil
}

func TestExtractor_WithMock(t *testing.T) {
	mockClient := &mockHTTPClient{
		responses: map[string]*http.Response{
			"https://example.com/album": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html>...</html>")),
			},
		},
	}

	// Inject mock
	extractor := &Extractor{
		client: mockClient,
	}

	result, err := extractor.Extract("https://example.com/album")
	// ... test with mocked HTTP
}
```

### Mock File System

```go
type mockFS struct {
	files map[string][]byte
}

func (m *mockFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func TestWithMockFS(t *testing.T) {
	fs := &mockFS{
		files: map[string][]byte{
			"config.yaml": []byte("token: test-token"),
		},
	}

	// Use mock in test
}
```

---

## Coverage

### Running Coverage

```bash
# Basic coverage
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./internal/domain
go test -cover ./internal/validation
```

### Coverage Goals

```bash
# Fail if coverage below 80%
go test -cover ./... | grep -E "coverage:.*[0-7][0-9]\.[0-9]%"
if [ $? -eq 0 ]; then
    echo "Coverage below 80%"
    exit 1
fi
```

### Viewing Coverage

```bash
# Generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

---

## Common Patterns

### Test Helpers

```go
// Helper functions simplify tests
func newTestTorrent(t *testing.T) *domain.Torrent {
	t.Helper()
	
	return &domain.Torrent{
		Title:        "Test Album",
		OriginalYear: 2020,
		Files: []domain.FileLike{
			&domain.Track{
				File:  domain.File{Path: "01.flac"},
				Disc:  1,
				Track: 1,
				Title: "Track",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
				},
			},
		},
	}
}

// Use in tests
func TestSomething(t *testing.T) {
	torrent := newTestTorrent(t)
	// Test with torrent
}
```

### Subtests

```go
func TestValidator_Rules(t *testing.T) {
	v := NewValidator()

	t.Run("composer not in title", func(t *testing.T) {
		torrent := newTestTorrent(t)
		// Modify torrent for this test
		
		issues := v.Rule_ComposerNotInTitle(torrent)
		// Check issues
	})

	t.Run("path length", func(t *testing.T) {
		torrent := newTestTorrent(t)
		// Modify torrent for this test
		
		issues := v.Rule_PathLength(torrent)
		// Check issues
	})
}
```

### Parallel Tests

```go
func TestExtractors(t *testing.T) {
	tests := []struct {
		name      string
		extractor Extractor
		url       string
	}{
		{"Discogs", NewDiscogsExtractor(), "https://discogs.com/..."},
		{"Naxos", NewNaxosExtractor(), "https://naxos.com/..."},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run in parallel
			
			result, err := tt.extractor.Extract(tt.url)
			// Test result
		})
	}
}
```

### Cleanup

```go
func TestWithCleanup(t *testing.T) {
	// Setup
	tmpDir := t.TempDir() // Auto-cleanup
	
	// Or manual cleanup
	t.Cleanup(func() {
		// Cleanup code runs even if test fails
		os.RemoveAll(tmpDir)
	})

	// Test code
}
```

### Testing Panics

```go
func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, got none")
		}
	}()

	// Code that should panic
	DoSomethingThatPanics()
}
```

---

## Best Practices

### 1. Use t.Helper()

```go
func assertEqual(t *testing.T, got, want string) {
	t.Helper() // Marks this as helper, improves error reporting
	
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

### 2. Test Both Success and Failure

```go
func TestParser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		result, err := Parse("valid input")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Check result
	})

	t.Run("failure", func(t *testing.T) {
		_, err := Parse("invalid input")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
```

### 3. Test Edge Cases

```go
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"single character", "a"},
		{"very long string", strings.Repeat("a", 10000)},
		{"special characters", "!@#$%^&*()"},
		{"unicode", "日本語"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with edge case input
		})
	}
}
```

### 4. Use Golden Files

```go
func TestOutput_Golden(t *testing.T) {
	output := GenerateOutput()
	
	goldenFile := "testdata/output.golden"
	
	if *update {
		// Update golden file
		os.WriteFile(goldenFile, []byte(output), 0644)
		t.Skip("Updated golden file")
	}
	
	expected, _ := os.ReadFile(goldenFile)
	if output != string(expected) {
		t.Errorf("Output differs from golden file")
	}
}

var update = flag.Bool("update", false, "update golden files")
```

### 5. Test Timeouts

```go
func TestWithTimeout(t *testing.T) {
	done := make(chan bool)
	
	go func() {
		// Long-running operation
		LongOperation()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Operation timed out")
	}
}
```

---

## Anti-Patterns

### ❌ Don't: Sleep in Tests

```go
// Bad
func TestBad(t *testing.T) {
	DoSomethingAsync()
	time.Sleep(100 * time.Millisecond) // Flaky!
}

// Good
func TestGood(t *testing.T) {
	done := make(chan bool)
	go func() {
		DoSomethingAsync()
		done <- true
	}()
	
	select {
	case <-done:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout")
	}
}
```

### ❌ Don't: Test Implementation Details

```go
// Bad - tests implementation
func TestBad(t *testing.T) {
	obj := NewObject()
	if obj.internalCounter != 0 { // Testing private field
		t.Error("counter should be 0")
	}
}

// Good - tests behavior
func TestGood(t *testing.T) {
	obj := NewObject()
	result := obj.GetCount() // Testing public API
	if result != 0 {
		t.Error("count should be 0")
	}
}
```

### ❌ Don't: Have Test Dependencies

```go
// Bad - tests depend on each other
func TestFirst(t *testing.T) {
	globalState = "value"
}

func TestSecond(t *testing.T) {
	// Depends on TestFirst running first
	if globalState != "value" {
		t.Error("wrong state")
	}
}

// Good - independent tests
func TestFirst(t *testing.T) {
	state := "value"
	// Use local state
}

func TestSecond(t *testing.T) {
	state := "value"
	// Independent setup
}
```

---

## Resources

- **Go Testing Package:** https://pkg.go.dev/testing
- **Table-Driven Tests:** https://go.dev/wiki/TableDrivenTests
- **Advanced Testing:** https://go.dev/blog/subtests
- **Test Fixtures:** https://go.dev/wiki/TestFixtures

---

**Last Updated:** 2025-01-XX  
**Version:** 1.0  
**Maintainer:** classical-tagger project
