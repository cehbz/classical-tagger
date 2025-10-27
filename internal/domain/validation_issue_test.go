package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidationIssue_DirectFieldAccess(t *testing.T) {
	issue := ValidationIssue{
		Level:   LevelError,
		Track:   1,
		Rule:    "2.3.16.4",
		Message: "Missing required tag 'Composer'",
	}

	if issue.Level != LevelError {
		t.Errorf("Level = %v, want %v", issue.Level, LevelError)
	}
	if issue.Track != 1 {
		t.Errorf("Track = %v, want %v", issue.Track, 1)
	}
	if issue.Rule != "2.3.16.4" {
		t.Errorf("Rule = %v, want %v", issue.Rule, "2.3.16.4")
	}
	if issue.Message != "Missing required tag 'Composer'" {
		t.Errorf("Message = %v, want %v", issue.Message, "Missing required tag 'Composer'")
	}
}

func TestValidationIssue_String(t *testing.T) {
	tests := []struct {
		Name    string
		Issue   ValidationIssue
		WantStr []string // substrings that should be present
	}{
		{
			Name: "track-level error",
			Issue: ValidationIssue{
				Level:   LevelError,
				Track:   3,
				Rule:    "2.3.16.4",
				Message: "Missing required tag 'Title'",
			},
			WantStr: []string{
				"ERROR",
				"Track 3",
				"2.3.16.4",
				"Missing required tag 'Title'",
			},
		},
		{
			Name: "album-level warning",
			Issue: ValidationIssue{
				Level:   LevelWarning,
				Track:   0,
				Rule:    "2.3.16.4",
				Message: "Missing recommended tag 'Year'",
			},
			WantStr: []string{
				"WARNING",
				"Album",
				"2.3.16.4",
				"Missing recommended tag 'Year'",
			},
		},
		{
			Name: "directory-level info",
			Issue: ValidationIssue{
				Level:   LevelInfo,
				Track:   -1,
				Rule:    "2.3.12",
				Message: "Consider shortening path",
			},
			WantStr: []string{
				"INFO",
				"Directory",
				"2.3.12",
				"Consider shortening path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := tt.Issue.String()
			for _, want := range tt.WantStr {
				if !strings.Contains(got, want) {
					t.Errorf("ValidationIssue.String() missing substring %q, got %q", want, got)
				}
			}
		})
	}
}

func TestValidationIssue_JSONSerialization(t *testing.T) {
	issue := ValidationIssue{
		Level:   LevelError,
		Track:   5,
		Rule:    "2.3.16.4",
		Message: "Missing composer",
	}

	// Marshal to JSON
	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded ValidationIssue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify round-trip
	if decoded.Level != issue.Level {
		t.Errorf("Level after round-trip = %v, want %v", decoded.Level, issue.Level)
	}
	if decoded.Track != issue.Track {
		t.Errorf("Track after round-trip = %v, want %v", decoded.Track, issue.Track)
	}
	if decoded.Rule != issue.Rule {
		t.Errorf("Rule after round-trip = %v, want %v", decoded.Rule, issue.Rule)
	}
	if decoded.Message != issue.Message {
		t.Errorf("Message after round-trip = %v, want %v", decoded.Message, issue.Message)
	}
}
