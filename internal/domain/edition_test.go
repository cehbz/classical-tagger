package domain

import "testing"

func TestNewEdition(t *testing.T) {
	tests := []struct {
		name        string
		label       string
		editionYear int
		wantErr     bool
	}{
		{
			name:        "valid edition",
			label:       "harmonia mundi",
			editionYear: 2013,
			wantErr:     false,
		},
		{
			name:        "empty label",
			label:       "",
			editionYear: 2013,
			wantErr:     true,
		},
		{
			name:        "whitespace label",
			label:       "   ",
			editionYear: 2013,
			wantErr:     true,
		},
		{
			name:        "zero year",
			label:       "Deutsche Grammophon",
			editionYear: 0,
			wantErr:     true,
		},
		{
			name:        "negative year",
			label:       "Decca",
			editionYear: -1,
			wantErr:     true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEdition(tt.label, tt.editionYear)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEdition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Label() != tt.label {
					t.Errorf("NewEdition().Label() = %v, want %v", got.Label(), tt.label)
				}
				if got.Year() != tt.editionYear {
					t.Errorf("NewEdition().Year() = %v, want %v", got.Year(), tt.editionYear)
				}
				if got.CatalogNumber() != "" {
					t.Errorf("NewEdition().CatalogNumber() = %v, want empty string", got.CatalogNumber())
				}
			}
		})
	}
}

func TestEdition_WithCatalogNumber(t *testing.T) {
	edition, _ := NewEdition("harmonia mundi", 2013)
	edition = edition.WithCatalogNumber("HMC902170")
	
	if got := edition.CatalogNumber(); got != "HMC902170" {
		t.Errorf("Edition.CatalogNumber() = %v, want %v", got, "HMC902170")
	}
}

func TestEdition_Validate(t *testing.T) {
	tests := []struct {
		name           string
		edition        Edition
		wantIssueCount int
		wantWarnings   bool
	}{
		{
			name: "complete edition",
			edition: func() Edition {
				e, _ := NewEdition("harmonia mundi", 2013)
				return e.WithCatalogNumber("HMC902170")
			}(),
			wantIssueCount: 0,
			wantWarnings:   false,
		},
		{
			name: "missing catalog number",
			edition: func() Edition {
				e, _ := NewEdition("harmonia mundi", 2013)
				return e
			}(),
			wantIssueCount: 1,
			wantWarnings:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := tt.edition.Validate()
			if len(issues) != tt.wantIssueCount {
				t.Errorf("Edition.Validate() returned %d issues, want %d", len(issues), tt.wantIssueCount)
			}
			if tt.wantWarnings && len(issues) > 0 {
				if issues[0].Level() != LevelWarning {
					t.Errorf("Edition.Validate() issue level = %v, want WARNING", issues[0].Level())
				}
			}
		})
	}
}
