package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/storage"
)

func TestValidateJSONFiles_ValidAlbum(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")

	// Create a valid album JSON
	album := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Test Label",
			CatalogNumber: "TL123",
			Year:          2013,
		},
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
					{Name: "Ensemble", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	// Save to JSON file
	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save test JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if report.Album == nil {
		t.Fatal("Album should be loaded")
	}

	if report.Album.Title != album.Title {
		t.Errorf("Album title = %q, want %q", report.Album.Title, album.Title)
	}

	if len(report.LoadErrors) > 0 {
		t.Errorf("Unexpected load errors: %v", report.LoadErrors)
	}
}

func TestValidateJSONFiles_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "invalid.json")

	// Create invalid JSON file
	if err := os.WriteFile(jsonFile, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if len(report.LoadErrors) == 0 {
		t.Error("Expected load error for invalid JSON")
	}

	if report.Album != nil {
		t.Error("Album should not be loaded when JSON is invalid")
	}
}

func TestValidateJSONFiles_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "nonexistent.json")

	// Validate non-existent file
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if len(report.LoadErrors) == 0 {
		t.Error("Expected load error for missing file")
	}

	if report.Album != nil {
		t.Error("Album should not be loaded when file is missing")
	}
}

func TestValidateJSONFiles_WithReference(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")
	refFile := filepath.Join(tmpDir, "reference.json")

	// Create album JSON
	album := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
				},
			},
		},
	}

	// Create reference JSON
	reference := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
				},
			},
		},
	}

	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save album JSON: %v", err)
	}
	if err := repo.SaveToFile(reference, refFile); err != nil {
		t.Fatalf("Failed to save reference JSON: %v", err)
	}

	// Validate with reference
	report, err := ValidateJSONFiles(jsonFile, refFile)
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if report.Album == nil {
		t.Fatal("Album should be loaded")
	}

	if len(report.LoadErrors) > 0 {
		t.Errorf("Unexpected load errors: %v", report.LoadErrors)
	}
}

func TestValidateJSONFiles_InvalidReference(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")
	refFile := filepath.Join(tmpDir, "invalid_ref.json")

	// Create valid album JSON
	album := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
				},
			},
		},
	}

	// Create invalid reference JSON
	if err := os.WriteFile(refFile, []byte("{ invalid }"), 0644); err != nil {
		t.Fatalf("Failed to create invalid reference: %v", err)
	}

	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save album JSON: %v", err)
	}

	// Validate with invalid reference
	report, err := ValidateJSONFiles(jsonFile, refFile)
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if report.Album == nil {
		t.Fatal("Album should be loaded")
	}

	// Should have load error for reference but continue validation
	if len(report.LoadErrors) == 0 {
		t.Error("Expected load error for invalid reference file")
	}
}

func TestValidationReport_HasErrors(t *testing.T) {
	report := &ValidationReport{
		Issues: []domain.ValidationIssue{
			{Level: domain.LevelWarning, Track: 1, Rule: "2.3.1", Message: "warning"},
			{Level: domain.LevelInfo, Track: 1, Rule: "2.3.1", Message: "info"},
		},
	}

	if report.HasErrors() {
		t.Error("Report should not have errors")
	}

	report.Issues = append(report.Issues, domain.ValidationIssue{
		Level: domain.LevelError, Track: 1, Rule: "2.3.1", Message: "error",
	})

	if !report.HasErrors() {
		t.Error("Report should have errors")
	}

	// Test with load errors
	report = &ValidationReport{
		LoadErrors: []error{os.ErrNotExist},
	}

	if !report.HasErrors() {
		t.Error("Report should have errors due to load errors")
	}
}

func TestValidationReport_HasWarnings(t *testing.T) {
	report := &ValidationReport{
		Issues: []domain.ValidationIssue{
			{Level: domain.LevelInfo, Track: 1, Rule: "2.3.1", Message: "info"},
		},
	}

	if report.HasWarnings() {
		t.Error("Report should not have warnings")
	}

	report.Issues = append(report.Issues, domain.ValidationIssue{
		Level: domain.LevelWarning, Track: 1, Rule: "2.3.1", Message: "warning",
	})

	if !report.HasWarnings() {
		t.Error("Report should have warnings")
	}
}

func TestValidateJSONFiles_ValidationIssues(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")

	// Create album with validation issues (missing composer, all caps title)
	album := &domain.Album{
		Title:        "ALL CAPS TITLE",
		OriginalYear: 2013,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "ALL CAPS TRACK",
				Artists: []domain.Artist{
					{Name: "Ensemble", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save test JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	// Should have validation issues
	if len(report.Issues) == 0 {
		t.Error("Expected validation issues")
	}

	// Should have errors (missing composer, capitalization issues)
	if !report.HasErrors() {
		t.Error("Report should have errors")
	}
}

func TestValidateJSONFiles_EmptyJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")

	// Create empty JSON object
	if err := os.WriteFile(jsonFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create empty JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	// Should load successfully (empty album is valid JSON)
	if report.Album == nil {
		t.Fatal("Album should be loaded even if empty")
	}

	// Should have validation issues for missing required fields
	if len(report.Issues) == 0 {
		t.Error("Expected validation issues for empty album")
	}
}

func TestValidateJSONFiles_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "malformed.json")

	// Create malformed JSON
	malformedJSON := `{
		"title": "Test",
		"tracks": [
			{
				"disc": 1,
				"track": 1,
				"title": "Track 1"
				// Missing closing brace
		]
	}`

	if err := os.WriteFile(jsonFile, []byte(malformedJSON), 0644); err != nil {
		t.Fatalf("Failed to create malformed JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	// Should have load error
	if len(report.LoadErrors) == 0 {
		t.Error("Expected load error for malformed JSON")
	}

	if report.Album != nil {
		t.Error("Album should not be loaded when JSON is malformed")
	}
}

func TestValidateJSONFiles_ValidMultiDiscAlbum(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")

	// Create valid multi-disc album
	album := &domain.Album{
		Title:        "Multi-Disc Album",
		OriginalYear: 2013,
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Disc 1 Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
					{Name: "Ensemble", Role: domain.RoleEnsemble},
				},
			},
			{
				Disc:  2,
				Track: 1,
				Title: "Disc 2 Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
					{Name: "Ensemble", Role: domain.RoleEnsemble},
				},
			},
		},
	}

	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save test JSON: %v", err)
	}

	// Validate
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if report.Album == nil {
		t.Fatal("Album should be loaded")
	}

	if len(report.Album.Tracks) != 2 {
		t.Errorf("Track count = %d, want 2", len(report.Album.Tracks))
	}

	if len(report.LoadErrors) > 0 {
		t.Errorf("Unexpected load errors: %v", report.LoadErrors)
	}
}

func TestValidateJSONFiles_JSONSerialization(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "album.json")

	// Create album and save it
	album := &domain.Album{
		Title:        "Test Album",
		OriginalYear: 2013,
		Edition: &domain.Edition{
			Label:         "Test Label",
			CatalogNumber: "TL123",
			Year:          2013,
		},
		Tracks: []*domain.Track{
			{
				Disc:  1,
				Track: 1,
				Title: "Track 1",
				Artists: []domain.Artist{
					{Name: "Composer", Role: domain.RoleComposer},
				},
			},
		},
	}

	repo := storage.NewRepository()
	if err := repo.SaveToFile(album, jsonFile); err != nil {
		t.Fatalf("Failed to save test JSON: %v", err)
	}

	// Read it back and verify it's valid JSON
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON is not valid: %v", err)
	}

	// Validate the file we just created
	report, err := ValidateJSONFiles(jsonFile, "")
	if err != nil {
		t.Fatalf("ValidateJSONFiles error: %v", err)
	}

	if report.Album == nil {
		t.Fatal("Album should be loaded")
	}

	if report.Album.Title != album.Title {
		t.Errorf("Album title = %q, want %q", report.Album.Title, album.Title)
	}
}
