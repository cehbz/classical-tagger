package domain

import "testing"

func TestEdition_Validate(t *testing.T) {
	tests := []struct {
		Name           string
		Edition        Edition
		WantIssueCount int
		WantWarnings   bool
	}{
		{
			Name: "complete edition",
			Edition: Edition{Label: "test label", Year: 2013, CatalogNumber: "HMC902170"},
			WantIssueCount: 0,
			WantWarnings:   false,
		},
		{
			Name: "missing catalog number",
			Edition: Edition{Label: "test label", Year: 2013},
			WantIssueCount: 1,
			WantWarnings:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			issues := tt.Edition.Validate()
			if len(issues) != tt.WantIssueCount {
				t.Errorf("Edition.Validate() returned %d issues, want %d", len(issues), tt.WantIssueCount)
			}
			if tt.WantWarnings && len(issues) > 0 {
				if issues[0].Level != LevelWarning {
					t.Errorf("Edition.Validate() issue level = %v, want WARNING", issues[0].Level)
				}
			}
		})
	}
}
