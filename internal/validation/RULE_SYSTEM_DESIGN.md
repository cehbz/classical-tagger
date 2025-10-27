# Rule System Design Document

## Overview

This document describes the "super magic" rule discovery system for classical music validation. The system automatically discovers validation rules using reflection, inspired by Go's `testing` package.

## Core Concept

**Rules are methods on a struct.** Any exported method on the `Rules` struct that matches a specific signature is automatically discovered and executed as a validation rule.

This eliminates manual bookkeeping - just write a method and it becomes a rule.

## Type Definitions

### RuleMetadata

Describes a rule's identity and severity:

```go
type RuleMetadata struct {
    id     string   // Rule section (e.g., "2.3.12", "classical.composer")
    name   string   // Human-readable name
    level  Level    // LevelError or LevelWarning
    weight float64  // For scoring (typically 1.0)
}
```

**Methods:**
- `ID() string`
- `Name() string`
- `Level() Level`
- `Weight() float64`
- `Pass() RuleResult` - Creates passing result
- `Fail(issues ...ValidationIssue) RuleResult` - Creates failing result

### RuleResult

What every rule method returns:

```go
type RuleResult struct {
    meta   RuleMetadata
    issues []ValidationIssue
}
```

**Methods:**
- `Meta RuleMetadata`
- `Issues() []ValidationIssue`
- `Passed() bool` - True if no issues

### Rules

The container for all rule methods:

```go
type Rules struct{}
```

**Methods:**
- `Discover() []RuleFunc` - Uses reflection to find all rule methods
- `PathLength(actual, reference *domain.AlbumMetadata) RuleResult` - Example rule
- `ComposerNotInTitle(actual, reference *domain.AlbumMetadata) RuleResult` - Example rule
- ... (one method per validation rule)

### RuleFunc

Type alias for rule function signature:

```go
type RuleFunc func(actual, reference *domain.AlbumMetadata) RuleResult
```

## Rule Discovery Algorithm

The `Discover()` method uses reflection to find valid rule methods:

1. **Iterate all methods** on `*Rules` type
2. **Filter by signature:**
   - Must be exported (uppercase first letter)
   - Must take exactly 2 parameters: `(*domain.AlbumMetadata, *domain.AlbumMetadata)`
   - Must return exactly 1 value: `RuleResult`
   - Must NOT be named "Discover" (skip the discovery method itself)
3. **Wrap each method** as a `RuleFunc` using reflection
4. **Return slice** of discovered rule functions

### Why This Works

Go's `reflect` package allows us to:
- Enumerate all methods on a type: `reflect.TypeOf(r).NumMethod()`
- Inspect method signatures: `method.Type.NumIn()`, `method.Type.In(i)`, `method.Type.Out(0)`
- Call methods dynamically: `methodValue.Call(args)`

This is the same mechanism `testing.T` uses to discover `Test*` functions.

## Rule Method Structure

Every rule method follows this pattern:

```go
func (r *Rules) RuleName(actual, reference *domain.AlbumMetadata) RuleResult {
    // 1. Define metadata
    meta := RuleMetadata{
        ID:     "section.number",
        Name:   "Human readable description",
        Level:  LevelError,  // or LevelWarning
        Weight: 1.0,
    }
    
    // 2. Collect issues
    var issues []ValidationIssue
    
    // 3. Perform validation checks
    for _, track := range actual.Tracks {
        if violatesRule(track) {
            issues = append(issues, ValidationIssue{
                Level: LevelError,
                Track: track.Track,
                Rule: meta.ID,
                Message: "Description of violation",
            })
        }
    }
    
    // 4. Return result
    if len(issues) == 0 {
        return meta.Pass
    }
    return RuleResult{Meta: meta, Issues: issues}
}
```

## Validation Runner

The runner executes all discovered rules and aggregates results:

```go
type ValidationResult struct {
    albumPath      string
    ruleResults    []RuleResult
    totalRules     int
    passedRules    int
    failedRules    int
    errorCount     int
    warningCount   int
}

func RunAll(actual, reference *domain.AlbumMetadata, rules []RuleFunc) *ValidationResult
```

**Algorithm:**
1. Create empty `ValidationResult`
2. For each rule function:
   - Call `rule(actual, reference)` to get `RuleResult`
   - Append to `ruleResults`
   - Update counters (passed/failed, errors/warnings)
3. Calculate improvement score
4. Return aggregated result

## Improvement Scoring

Score = 1.0 - (total penalty / max possible penalty)

Where:
- Each failed rule contributes its `weight` to penalty
- Passed rules contribute 0 penalty
- Score of 1.0 = perfect compliance
- Score of 0.0 = all rules failed

```go
func (r *ValidationResult) ImprovementScore() float64 {
    maxPenalty := 0.0
    for _, rr := range r.ruleResults {
        maxPenalty += rr.Meta.Weight
    }
    
    actualPenalty := 0.0
    for _, rr := range r.ruleResults {
        if !rr.Passed() {
            actualPenalty += rr.Meta.Weight
        }
    }
    
    if maxPenalty == 0 {
        return 0
    }
    
    return math.Max(0, 1.0 - (actualPenalty / maxPenalty))
}
```

## File Organization

```
internal/validation/
├── rule.go                      # RuleMetadata, RuleResult types
├── rules.go                     # Rules struct, Discover() method
├── runner.go                    # RunAll(), ValidationResult
├── rule_path_length.go          # PathLength() rule method
├── rule_composer_in_title.go   # ComposerNotInTitle() rule method
├── rule_track_sort.go           # TrackSort() rule method
├── rule_title_accuracy.go      # TitleAccuracy() rule method
├── rule_*.go                    # One file per rule
└── rules_test.go                # Tests for discovery and execution
```

## Adding a New Rule

1. Create `internal/validation/rule_my_check.go`
2. Add method to `Rules`:

```go
func (r *Rules) MyCheck(actual, reference *domain.AlbumMetadata) RuleResult {
    meta := RuleMetadata{
        ID:     "2.3.X",
        Name:   "My validation check",
        Level:  LevelError,
        Weight: 1.0,
    }
    
    // ... validation logic ...
    
    if len(issues) == 0 {
        return meta.Pass
    }
    return RuleResult{Meta: meta, Issues: issues}
}
```

3. Done! Automatically discovered and executed.

## Removing a Rule

1. Delete the method from `Rules`
2. Done!

## Testing Strategy

### Unit Tests for Individual Rules

Each rule should have its own test file:

```go
// In validation/rule_path_length_test.go
func TestRules_PathLength(t *testing.T) {
    rules := NewRules()
    
    tests := []struct {
        name       string
        album      *domain.AlbumMetadata
        wantPass   bool
        wantIssues int
    }{
        // Test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.Name, func(t *testing.T) {
            result := rules.PathLength(tt.album, tt.album)
            // Assert result...
        })
    }
}
```

### Discovery Tests

Verify the discovery mechanism works:

```go
func TestRules_Discover(t *testing.T) {
    rules := NewRules()
    discovered := rules.Discover()
    
    // Should find multiple rules
    if len(discovered) == 0 {
        t.Fatal("No rules discovered")
    }
    
    // All rules should be executable
    empty := &domain.AlbumMetadata{}
    for _, rule := range discovered {
        result := rule(empty, empty)
        // Verify result structure...
    }
}
```

### Integration Tests

Test the full validation flow:

```go
func TestValidationFlow(t *testing.T) {
    actual := loadTestAlbum(t)
    reference := loadReferenceAlbum(t)
    
    rules := All()
    result := RunAll(actual, reference, rules)
    
    // Verify result structure, scoring, etc.
}
```

## Advantages

1. **Zero bookkeeping** - No manual registration
2. **Self-documenting** - All rules visible as methods on one type
3. **Easy to add/remove** - Just add/delete methods
4. **Familiar pattern** - Works like `testing.T`
5. **Type-safe at runtime** - Discovery validates signatures
6. **Easy to test** - Call methods directly in unit tests
7. **Easy to debug** - Each rule is a normal Go method

## Disadvantages

1. **Runtime reflection** - Slight startup cost (negligible)
2. **Stack traces** - Show `reflect.Call` in errors (minor)
3. **Signature validation at runtime** - Not compile-time (caught by tests)
4. **"Magic"** - Requires understanding reflection (standard Go pattern)

## Design Decisions

### Why Methods Instead of Functions?

**Methods allow:**
- Grouping related rules under one type
- Potential for shared helper methods
- Clear namespace (Rules.PathLength vs pathLengthRule)
- Future extension (could add state if needed)

### Why Not Manual Registration?

**Auto-discovery eliminates:**
- Forgetting to register a rule
- Keeping registry in sync with actual rules
- Boilerplate `init()` functions
- Import side effects

### Why Require Reference Parameter?

**Always requiring reference simplifies:**
- Signature checking (one less case)
- API consistency (every rule takes same params)
- Future flexibility (can add rules that need it)

Rules that don't need reference just ignore it.

## Comparison to Go Testing

| Aspect | Go Testing | Our Validation |
|--------|-----------|----------------|
| Container | `*testing.T` | `*Rules` |
| Method Pattern | `Test*` prefix | Match signature |
| Discovery | Package-level funcs | Methods on struct |
| Execution | `go test` | `RunAll()` |
| Result | Pass/Fail/Skip | Pass/Fail with issues |
| Reporting | Text output | Structured result |

Both use reflection to discover and execute code automatically.

## Future Extensions

### Parallel Execution

Rules are independent and could run in parallel:

```go
func (r *Rules) Discover() []RuleFunc {
    // ... discovery logic ...
    // Could add parallel execution hints
}
```

### Rule Dependencies

If needed, could add metadata for rule ordering:

```go
type RuleMetadata struct {
    // ...
    dependsOn []string  // Rule IDs that must pass first
}
```

### Rule Categories

Could filter rules by category:

```go
func (r *Rules) DiscoverStructural() []RuleFunc {
    all := r.Discover()
    // Filter to structural rules only
}
```

### Custom Comparators

Rules that need specialized comparison logic:

```go
func (r *Rules) TitleAccuracy(actual, reference *domain.AlbumMetadata) RuleResult {
    // Use fuzzy string matching, normalization, etc.
}
```

## Conclusion

This design provides a clean, maintainable, and extensible validation system with minimal boilerplate. The "magic" is standard Go reflection, familiar to anyone who has used the testing package.

The key insight: **Rules are just methods that follow a convention.** Everything else follows from that.
