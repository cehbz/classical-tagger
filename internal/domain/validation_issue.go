package domain

import "fmt"

// ValidationIssue represents a single validation problem.
// It is immutable after creation.
type ValidationIssue struct {
	level   Level
	track   int    // 0 for album-level, -1 for directory-level, >0 for track number
	rule    string // section number from rules (e.g., "2.3.16.4")
	message string // context-specific message
}

// NewIssue creates a new ValidationIssue.
func NewIssue(level Level, track int, rule, message string) ValidationIssue {
	return ValidationIssue{
		level:   level,
		track:   track,
		rule:    rule,
		message: message,
	}
}

// Level returns the severity level of the issue.
func (v ValidationIssue) Level() Level {
	return v.level
}

// Track returns the track number, 0 for album-level, -1 for directory-level.
func (v ValidationIssue) Track() int {
	return v.track
}

// Rule returns the rule section number.
func (v ValidationIssue) Rule() string {
	return v.rule
}

// Message returns the context-specific message.
func (v ValidationIssue) Message() string {
	return v.message
}

// String returns a formatted string representation of the issue.
func (v ValidationIssue) String() string {
	var location string
	switch {
	case v.track > 0:
		location = fmt.Sprintf("Track %d", v.track)
	case v.track == 0:
		location = "Album"
	case v.track == -1:
		location = "Directory"
	}
	
	return fmt.Sprintf("[%s] %s: %s - %s", v.level, location, v.rule, v.message)
}
