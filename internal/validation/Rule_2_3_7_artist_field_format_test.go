package validation

import (
	"testing"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_ArtistFieldFormat(t *testing.T) {
	rules := NewRules()
	
	tests := []struct {
		name         string
		actual       *domain.Album
		reference    *domain.Album
		wantPass     bool
		wantWarnings int
		wantInfo     int
	}{
		{
			name: "valid - has performers",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
				"Berlin Phil", domain.RoleEnsemble,
			),
			wantPass: true,
		},
		{
			name: "warning - only composer",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
			),
			wantPass:     false,
			wantWarnings: 1,
		},
		{
			name: "valid - just performers (no composer)",
			actual: buildAlbumWithArtists(
				"Pollini", domain.RoleSoloist,
				"Berlin Phil", domain.RoleEnsemble,
			),
			wantPass: true,
		},
		{
			name: "valid - ensemble only",
			actual: buildAlbumWithArtists(
				"Beethoven", domain.RoleComposer,
				"Emerson Quartet", domain.RoleEnsemble,
			),
			wantPass: true,
		},
		{
			name: "info - performer count differs from reference",
			actual: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
			),
			reference: buildAlbumWithArtists(
				"Bach", domain.RoleComposer,
				"Pollini", domain.RoleSoloist,
				"Orchestra", domain.RoleEnsemble,
			),
			wantPass: false,
			wantInfo: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.ArtistFieldFormat(tt.actual, tt.reference)
			
			if result.Passed() != tt.wantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.wantPass)
			}
			
			if !tt.wantPass {
				warningCount := 0
				infoCount := 0
				for _, issue := range result.Issues() {
					if issue.Level() == domain.LevelWarning {
						warningCount++
					} else if issue.Level() == domain.LevelInfo {
						infoCount++
					}
				}
				
				if tt.wantWarnings > 0 && warningCount != tt.wantWarnings {
					t.Errorf("Warnings = %d, want %d", warningCount, tt.wantWarnings)
				}
				if tt.wantInfo > 0 && infoCount != tt.wantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.wantInfo)
				}
				
				for _, issue := range result.Issues() {
					t.Logf("  Issue [%s]: %s", issue.Level(), issue.Message())
				}
			}
		})
	}
}

func TestGetPerformers(t *testing.T) {
	composer, _ := domain.NewArtist("Beethoven", domain.RoleComposer)
	soloist, _ := domain.NewArtist("Pollini", domain.RoleSoloist)
	ensemble, _ := domain.NewArtist("Orchestra", domain.RoleEnsemble)
	arranger, _ := domain.NewArtist("Mahler", domain.RoleArranger)
	
	tests := []struct {
		name    string
		artists []domain.Artist
		want    []string
	}{
		{
			name:    "all roles",
			artists: []domain.Artist{composer, soloist, ensemble, arranger},
			want:    []string{"Pollini", "Orchestra"},
		},
		{
			name:    "only composer",
			artists: []domain.Artist{composer},
			want:    []string{},
		},
		{
			name:    "only performers",
			artists: []domain.Artist{soloist, ensemble},
			want:    []string{"Pollini", "Orchestra"},
		},
		{
			name:    "empty",
			artists: []domain.Artist{},
			want:    []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPerformers(tt.artists)
			if len(got) != len(tt.want) {
				t.Errorf("getPerformers() count = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getPerformers()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
