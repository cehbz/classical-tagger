package validation

import (
	"testing"

	"github.com/cehbz/classical-tagger/internal/domain"
)

func TestRules_AlbumArtistTag(t *testing.T) {
	rules := NewRules()

	tests := []struct {
		Name     string
		Torrent *domain.Torrent
		WantPass bool
		WantInfo int
	}{
		{
			Name: "pass - single track",
			Torrent: &domain.Torrent{
				Title:        "Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Pollini", Role: domain.RoleSoloist},
						},
					},
				},
			},
			WantPass: true,
		},
		{
			Name: "info - dominant performer",
			Torrent: &domain.Torrent{
				Title:        "Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name: "info - various composers",
			Torrent: &domain.Torrent{
				Title:        "Various Artists Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: false,
			WantInfo: 1,
		},
		{
			Name: "pass - no dominant performer",
			Torrent: &domain.Torrent{
				Title:        "Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Pollini", Role: domain.RoleSoloist},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 2,
						Title: "Work 2",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Arrau", Role: domain.RoleSoloist},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 3,
						Title: "Work 3",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Brendel", Role: domain.RoleSoloist},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 4,
						Title: "Work 4",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Ashkenazy", Role: domain.RoleSoloist},
						},
					},
				},
			},
			WantPass: true,
		},
		{
			Name: "pass - two composers only",
				Torrent: &domain.Torrent{
				Title:        "Album",
				OriginalYear: 1963,
				Files: []domain.FileLike{
					&domain.Track{
						Disc:  1,
						Track: 1,
						Title: "Work 1",
						Artists: []domain.Artist{
							domain.Artist{Name: "Beethoven", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
					&domain.Track{
						Disc:  1,
						Track: 2,
						Title: "Work 2",
						Artists: []domain.Artist{
							domain.Artist{Name: "Mozart", Role: domain.RoleComposer},
							domain.Artist{Name: "Orchestra", Role: domain.RoleEnsemble},
						},
					},
				},
			},
			WantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := rules.AlbumArtistTag(tt.Torrent, nil)

			if result.Passed() != tt.WantPass {
				t.Errorf("Passed = %v, want %v", result.Passed(), tt.WantPass)
			}

			if !tt.WantPass {
				infoCount := 0
				for _, issue := range result.Issues {
					if issue.Level == domain.LevelInfo {
						infoCount++
					}
				}

				if infoCount != tt.WantInfo {
					t.Errorf("Info = %d, want %d", infoCount, tt.WantInfo)
				}

				for _, issue := range result.Issues {
					t.Logf("  Issue [%s]: %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

func TestRules_AlbumArtist_LaxInclusion(t *testing.T) {
	rules := NewRules()

	// AlbumArtist two names; only one appears on some tracks -> pass (lax inclusion requires at least once across torrent)
	torrent := &domain.Torrent{
		Title:        "Album",
		OriginalYear: 2000,
		AlbumArtist: []domain.Artist{
			{Name: "RIAS-Kammerchor", Role: domain.RoleUnknown},
			{Name: "Hans-Christoph Rademann", Role: domain.RoleUnknown},
		},
		Files: []domain.FileLike{
			&domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}, {Name: "RIAS-Kammerchor", Role: domain.RoleUnknown}}},
			&domain.Track{Disc: 1, Track: 2, Title: "Work 2", Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}}},
			&domain.Track{Disc: 1, Track: 3, Title: "Work 3", Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}, {Name: "Hans-Christoph Rademann", Role: domain.RoleUnknown}}},
		},
	}
	res := rules.AlbumArtistTag(torrent, nil)
	if !res.Passed() {
		t.Errorf("Expected pass when AlbumArtist names appear at least once across tracks")
	}

	// Missing one AlbumArtist entirely -> error
	torrentMissing := &domain.Torrent{
		Title:        "Album",
		OriginalYear: 2000,
		AlbumArtist: []domain.Artist{
			{Name: "RIAS-Kammerchor", Role: domain.RoleUnknown},
			{Name: "Hans-Christoph Rademann", Role: domain.RoleUnknown},
		},
		Files: []domain.FileLike{
			&domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}, {Name: "RIAS-Kammerchor", Role: domain.RoleUnknown}}},
		},
	}
	res2 := rules.AlbumArtistTag(torrentMissing, nil)
	if res2.Passed() {
		t.Errorf("Expected failure when an AlbumArtist name is missing from all tracks")
	}
}

func TestRules_AlbumArtist_InclusionInvariant(t *testing.T) {
	rules := NewRules()

	albumArtist := []domain.Artist{
		{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
		{Name: "Herbert von Karajan", Role: domain.RoleConductor},
	}

	// Missing inclusion on track -> expect error(s)
	torrentMissing := &domain.Torrent{
		Title:        "Album",
		OriginalYear: 1977,
		AlbumArtist:  albumArtist,
		Files: []domain.FileLike{
			&domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{{Name: "Beethoven", Role: domain.RoleComposer}}},
		},
	}
	resMissing := rules.AlbumArtistTag(torrentMissing, nil)
	if resMissing.Passed() {
		t.Errorf("Expected failure when AlbumArtist not included in track artists")
	}

	// Inclusion present on track -> expect pass
	torrentIncluded := &domain.Torrent{
		Title:        "Album",
		OriginalYear: 1977,
		AlbumArtist:  albumArtist,
		Files: []domain.FileLike{
			&domain.Track{Disc: 1, Track: 1, Title: "Work 1", Artists: []domain.Artist{
				{Name: "Beethoven", Role: domain.RoleComposer},
				{Name: "Berlin Philharmonic", Role: domain.RoleEnsemble},
				{Name: "Herbert von Karajan", Role: domain.RoleConductor},
			}},
		},
	}
	resIncluded := rules.AlbumArtistTag(torrentIncluded, nil)
	if !resIncluded.Passed() {
		t.Errorf("Expected pass when AlbumArtist is included in track artists")
	}

	// Various Artists should not require inclusion
	va := &domain.Torrent{
		Title:        "Various Artists Sampler",
		OriginalYear: 2001,
		AlbumArtist:  []domain.Artist{{Name: "Various Artists", Role: domain.RoleEnsemble}},
		Files: []domain.FileLike{
			&domain.Track{Disc: 1, Track: 1, Title: "Track", Artists: []domain.Artist{{Name: "Artist A", Role: domain.RoleSoloist}}},
		},
	}
	resVA := rules.AlbumArtistTag(va, nil)
	if !resVA.Passed() {
		t.Errorf("Expected pass for Various Artists without inclusion requirement")
	}
}
