package domain

import "fmt"

// ValidationIssue represents a single validation problem.
// Uses exported fields for direct access and JSON serialization.
type ValidationIssue struct {
	Level    Level  `json:"level"`             // Severity level (ERROR, WARNING, INFO)
	Required bool   `json:"required"`          // Whether the issue is required (e.g., for extraction)
	Track    int    `json:"track"`             // 0 for album-level, -1 for directory-level, >0 for track number
	Rule     string `json:"rule"`              // Section number from rules (e.g., "2.3.16.4")
	Message  string `json:"message,omitempty"` // Context-specific message
}

// String returns a formatted string representation of the issue.
func (v ValidationIssue) String() string {
	var location string
	switch {
	case v.Track > 0:
		location = fmt.Sprintf("Track %d", v.Track)
	case v.Track == 0:
		location = "Album"
	case v.Track == -1:
		location = "Directory"
	}

	return fmt.Sprintf("[%s] %s: %s - %s", v.Level, location, v.Rule, v.Message)
}
