package domain

import (
	"strings"
	"testing"
)

func TestNewIssue(t *testing.T) {
	issue := NewIssue(LevelError, 1, "2.3.16.4", "Missing required tag 'Composer'")
	
	if issue.Level() != LevelError {
		t.Errorf("NewIssue().Level() = %v, want %v", issue.Level(), LevelError)
	}
	if issue.Track() != 1 {
		t.Errorf("NewIssue().Track() = %v, want %v", issue.Track(), 1)
	}
	if issue.Rule() != "2.3.16.4" {
		t.Errorf("NewIssue().Rule() = %v, want %v", issue.Rule(), "2.3.16.4")
	}
	if issue.Message() != "Missing required tag 'Composer'" {
		t.Errorf("NewIssue().Message() = %v, want %v", issue.Message(), "Missing required tag 'Composer'")
	}
}

func TestValidationIssue_String(t *testing.T) {
	tests := []struct {
		name    string
		issue   ValidationIssue
		wantStr []string // substrings that should be present
	}{
		{
			name:  "track-level error",
			issue: NewIssue(LevelError, 3, "2.3.16.4", "Missing required tag 'Title'"),
			wantStr: []string{
				"ERROR",
				"Track 3",
				"2.3.16.4",
				"Missing required tag 'Title'",
			},
		},
		{
			name:  "album-level warning",
			issue: NewIssue(LevelWarning, 0, "2.3.16.4", "Missing recommended tag 'Year'"),
			wantStr: []string{
				"WARNING",
				"Album",
				"2.3.16.4",
				"Missing recommended tag 'Year'",
			},
		},
		{
			name:  "directory-level info",
			issue: NewIssue(LevelInfo, -1, "2.3.12", "Consider shortening path"),
			wantStr: []string{
				"INFO",
				"Directory",
				"2.3.12",
				"Consider shortening path",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.String()
			for _, want := range tt.wantStr {
				if !strings.Contains(got, want) {
					t.Errorf("ValidationIssue.String() missing substring %q, got %q", want, got)
				}
			}
		})
	}
}
