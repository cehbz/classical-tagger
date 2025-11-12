package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTorrent_Tracks(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []*Track
		wantLen int
	}{
		{
			name: "only tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{Track: 1, Title: "Track 1"},
					&Track{Track: 2, Title: "Track 2"},
					&Track{Track: 3, Title: "Track 3"},
				},
			},
			wantLen: 3,
		},
		{
			name: "mixed files and tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&File{Path: "cover.jpg"},
					&Track{Track: 1, Title: "Track 1"},
					&File{Path: "booklet.pdf"},
					&Track{Track: 2, Title: "Track 2"},
				},
			},
			wantLen: 2,
		},
		{
			name: "only files",
			torrent: &Torrent{
				Files: []FileLike{
					&File{Path: "cover.jpg"},
					&File{Path: "booklet.pdf"},
				},
			},
			wantLen: 0,
		},
		{
			name: "empty files",
			torrent: &Torrent{
				Files: []FileLike{},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.Tracks()
			if len(got) != tt.wantLen {
				t.Errorf("Tracks() returned %d tracks, want %d", len(got), tt.wantLen)
			}
			// Verify all returned items are tracks
			for i, track := range got {
				if track == nil {
					t.Errorf("Tracks()[%d] is nil", i)
				}
			}
		})
	}
}

func TestTorrent_IsMultiDisc(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    bool
	}{
		{
			name: "single disc",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{Disc: 1, Track: 1},
					&Track{Disc: 1, Track: 2},
					&Track{Disc: 1, Track: 3},
				},
			},
			want: false,
		},
		{
			name: "multi-disc with disc > 1",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{Disc: 1, Track: 1},
					&Track{Disc: 2, Track: 1},
					&Track{Disc: 2, Track: 2},
				},
			},
			want: true,
		},
		{
			name: "multi-disc with multiple disc numbers",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{Disc: 1, Track: 1},
					&Track{Disc: 3, Track: 1},
				},
			},
			want: true,
		},
		{
			name: "no tracks",
			torrent: &Torrent{
				Files: []FileLike{},
			},
			want: false,
		},
		{
			name: "only files, no tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&File{Path: "cover.jpg"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.IsMultiDisc()
			if got != tt.want {
				t.Errorf("IsMultiDisc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTorrent_AlbumArtists(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []Artist
	}{
		{
			name: "all tracks have same performer",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Johann Sebastian Bach", Role: RoleSoloist},
						},
					},
				},
			},
			want: []Artist{{Name: "Johann Sebastian Bach", Role: RoleSoloist}},
		},
		{
			name: "not all tracks have same performer",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleSoloist},
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Beethoven", Role: RoleSoloist},
							{Name: "Gould", Role: RoleSoloist},
						},
					},
				},
			},
			want: []Artist{
				{Name: "Gould", Role: RoleSoloist},
			},
		},
		{
			name: "multiple artists in all tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
							{Name: "Gould", Role: RoleSoloist},
							{Name: "Orchestra", Role: RoleEnsemble},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
							{Name: "Gould", Role: RoleSoloist},
							{Name: "Orchestra", Role: RoleEnsemble},
						},
					},
				},
			},
			want: []Artist{
				{Name: "Gould", Role: RoleSoloist},
				{Name: "Orchestra", Role: RoleEnsemble},
				// Non-performers are skipped by IsPerformer() check
			},
		},
		{
			name: "no tracks",
			torrent: &Torrent{
				Files: []FileLike{},
			},
			want: []Artist{},
		},
		{
			name: "non-performers are skipped",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Gould", Role: RoleComposer},
						},
					},
				},
			},
			want: []Artist{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.AlbumArtists()
			if len(got) != len(tt.want) {
				t.Errorf("AlbumArtists() returned %d artists, want %d", len(got), len(tt.want))
				return
			}
			// Check that all expected artists are present
			for _, wantArtist := range tt.want {
				found := false
				for _, gotArtist := range got {
					if gotArtist.Name == wantArtist.Name && gotArtist.Role == wantArtist.Role {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("AlbumArtists() missing artist: %+v", wantArtist)
				}
			}
		})
	}
}

func TestTorrent_Composers(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []string
	}{
		{
			name: "single composer",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Johann Sebastian Bach", Role: RoleComposer},
				},
			},
			want: []string{"Johann Sebastian Bach"},
		},
		{
			name: "multiple composers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Bach", Role: RoleComposer},
					{Name: "Mozart", Role: RoleComposer},
				},
			},
			want: []string{"Bach", "Mozart"},
		},
		{
			name: "composers and performers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Bach", Role: RoleComposer},
					{Name: "Gould", Role: RoleSoloist},
					{Name: "Mozart", Role: RoleComposer},
				},
			},
			want: []string{"Bach", "Mozart"},
		},
		{
			name: "no composers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Gould", Role: RoleSoloist},
					{Name: "Orchestra", Role: RoleEnsemble},
				},
			},
			want: []string{}, // Empty slice, not nil
		},
		{
			name: "empty album artist",
			torrent: &Torrent{
				AlbumArtist: []Artist{},
			},
			want: []string{}, // Empty slice, not nil
		},
		{
			name: "empty name composer",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "", Role: RoleComposer},
					{Name: "Bach", Role: RoleComposer},
				},
			},
			want: []string{"Bach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.Composers()
			// Compare lengths and contents, treating nil and empty slice as equivalent
			if len(got) != len(tt.want) {
				t.Errorf("Composers() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i, want := range tt.want {
				if i >= len(got) || got[i] != want {
					t.Errorf("Composers()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestTorrent_Performers(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []string
	}{
		{
			name: "single performer",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Glenn Gould", Role: RoleSoloist},
				},
			},
			want: []string{"Glenn Gould"},
		},
		{
			name: "multiple performers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Gould", Role: RoleSoloist},
					{Name: "Orchestra", Role: RoleEnsemble},
					{Name: "Conductor", Role: RoleConductor},
				},
			},
			want: []string{"Gould", "Orchestra", "Conductor"},
		},
		{
			name: "composers and performers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Bach", Role: RoleComposer},
					{Name: "Gould", Role: RoleSoloist},
					{Name: "Mozart", Role: RoleComposer},
				},
			},
			want: []string{"Gould"},
		},
		{
			name: "no performers",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Bach", Role: RoleComposer},
					{Name: "Mozart", Role: RoleComposer},
				},
			},
			want: []string{},
		},
		{
			name: "empty album artist",
			torrent: &Torrent{
				AlbumArtist: []Artist{},
			},
			want: []string{},
		},
		{
			name: "arranger is included as performer",
			torrent: &Torrent{
				AlbumArtist: []Artist{
					{Name: "Arranger", Role: RoleArranger},
					{Name: "Gould", Role: RoleSoloist},
				},
			},
			want: []string{"Arranger", "Gould"}, // Current implementation includes arrangers (checks != RoleComposer)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.Performers()
			// Compare lengths and contents, treating nil and empty slice as equivalent
			if len(got) != len(tt.want) {
				t.Errorf("Performers() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i, want := range tt.want {
				if i >= len(got) || got[i] != want {
					t.Errorf("Performers()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestTorrent_PrimaryComposers(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []string
	}{
		{
			name: "single composer on all tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
				},
			},
			want: []string{"Bach"},
		},
		{
			name: "composer on majority of tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Mozart", Role: RoleComposer},
						},
					},
				},
			},
			want: []string{"Bach"},
		},
		{
			name: "no composer on majority",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Mozart", Role: RoleComposer},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Beethoven", Role: RoleComposer},
						},
					},
				},
			},
			want: []string{}, // No composer on >50% of tracks (each on 1/3 = 33%)
		},
		{
			name: "multiple composers on majority",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Mozart", Role: RoleComposer},
						},
					},
					&Track{
						Track: 4,
						Artists: []Artist{
							{Name: "Mozart", Role: RoleComposer},
						},
					},
					&Track{
						Track: 5,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
							{Name: "Mozart", Role: RoleComposer},
						},
					},
				},
			},
			want: []string{"Bach", "Mozart"}, // Both on > 50%
		},
		{
			name: "no tracks",
			torrent: &Torrent{
				Files: []FileLike{},
			},
			want: []string{},
		},
		{
			name: "no composers",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.PrimaryComposers()
			// Compare lengths and contents, treating nil and empty slice as equivalent
			if len(got) != len(tt.want) {
				t.Errorf("PrimaryComposers() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i, want := range tt.want {
				if i >= len(got) || got[i] != want {
					t.Errorf("PrimaryComposers()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestTorrent_PrimaryPerformers(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		want    []string
	}{
		{
			name: "single performer on all tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
				},
			},
			want: []string{"Gould"},
		},
		{
			name: "performer on majority of tracks",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Pollini", Role: RoleSoloist},
						},
					},
				},
			},
			want: []string{"Gould"},
		},
		{
			name: "no performer on majority",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Pollini", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 3,
						Artists: []Artist{
							{Name: "Arrau", Role: RoleSoloist},
						},
					},
				},
			},
			want: []string{}, // No performer on >50% of tracks
		},
		{
			name: "multiple performers",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
							{Name: "Orchestra", Role: RoleEnsemble},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Gould", Role: RoleSoloist},
							{Name: "Orchestra", Role: RoleEnsemble},
						},
					},
				},
			},
			want: []string{"Gould", "Orchestra"},
		},
		{
			name: "composers are not performers",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
							{Name: "Gould", Role: RoleSoloist},
						},
					},
					&Track{
						Track: 2,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
							{Name: "Gould", Role: RoleSoloist},
						},
					},
				},
			},
			want: []string{"Gould"},
		},
		{
			name: "no tracks",
			torrent: &Torrent{
				Files: []FileLike{},
			},
			want: []string{},
		},
		{
			name: "no performers",
			torrent: &Torrent{
				Files: []FileLike{
					&Track{
						Track: 1,
						Artists: []Artist{
							{Name: "Bach", Role: RoleComposer},
						},
					},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.torrent.PrimaryPerformers()
			// Compare lengths and contents, treating nil and empty slice as equivalent
			if len(got) != len(tt.want) {
				t.Errorf("PrimaryPerformers() = %v (len=%d), want %v (len=%d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i, want := range tt.want {
				if i >= len(got) || got[i] != want {
					t.Errorf("PrimaryPerformers()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

func TestTorrent_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		torrent *Torrent
		wantErr bool
	}{
		{
			name: "basic torrent",
			torrent: &Torrent{
				RootPath:     "test/path",
				Title:        "Test Album",
				OriginalYear: 2020,
				Files: []FileLike{
					&Track{Track: 1, Title: "Track 1"},
					&Track{Track: 2, Title: "Track 2"},
				},
			},
			wantErr: false,
		},
		{
			name: "torrent with mixed files",
			torrent: &Torrent{
				RootPath: "test/path",
				Title:    "Test Album",
				Files: []FileLike{
					&File{Path: "cover.jpg"},
					&Track{Track: 1, Title: "Track 1"},
					&File{Path: "booklet.pdf"},
				},
			},
			wantErr: false,
		},
		{
			name: "torrent with edition",
			torrent: &Torrent{
				Title:        "Test Album",
				OriginalYear: 2020,
				Edition: &Edition{
					Label:         "Test Label",
					CatalogNumber: "CAT-123",
					Year:          2020,
				},
			},
			wantErr: false,
		},
		{
			name: "empty torrent",
			torrent: &Torrent{
				Title: "Empty Album",
				Files: []FileLike{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.torrent)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// Verify it can be unmarshaled back
				var unmarshaled Torrent
				if err := json.Unmarshal(data, &unmarshaled); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
			}
		})
	}
}

func TestTorrent_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *Torrent
		wantErr bool
	}{
		{
			name: "basic torrent",
			json: `{
				"root_path": "test/path",
				"title": "Test Album",
				"original_year": 2020,
				"files": [
					{"disc": 1, "track": 1, "title": "Track 1", "path": "01-Track 1.flac"},
					{"disc": 1, "track": 2, "title": "Track 2", "path": "02-Track 2.flac"}
				]
			}`,
			want: &Torrent{
				RootPath:     "test/path",
				Title:        "Test Album",
				OriginalYear: 2020,
			},
			wantErr: false,
		},
		{
			name: "torrent with files",
			json: `{
				"title": "Test Album",
				"files": [
					{"path": "cover.jpg"}
				]
			}`,
			want: &Torrent{
				Title: "Test Album",
			},
			wantErr: false,
		},
		{
			name: "torrent with empty files",
			json: `{
				"title": "Test Album",
				"files": []
			}`,
			want: &Torrent{
				Title: "Test Album",
				Files: []FileLike(nil),
			},
			wantErr: false,
		},
		{
			name: "torrent without files field",
			json: `{
				"title": "Test Album"
			}`,
			want: &Torrent{
				Title: "Test Album",
				Files: []FileLike(nil),
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Torrent
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.want != nil {
				if got.Title != tt.want.Title {
					t.Errorf("UnmarshalJSON() Title = %v, want %v", got.Title, tt.want.Title)
				}
				if got.RootPath != tt.want.RootPath {
					t.Errorf("UnmarshalJSON() RootPath = %v, want %v", got.RootPath, tt.want.RootPath)
				}
				if got.OriginalYear != tt.want.OriginalYear {
					t.Errorf("UnmarshalJSON() OriginalYear = %v, want %v", got.OriginalYear, tt.want.OriginalYear)
				}
			}
		})
	}
}

func TestTorrent_Save(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		torrent *Torrent
		wantErr bool
	}{
		{
			name: "save basic torrent",
			torrent: &Torrent{
				Title:        "Test Album",
				OriginalYear: 2020,
				Files: []FileLike{
					&Track{Track: 1, Title: "Track 1"},
				},
			},
			wantErr: false,
		},
		{
			name: "save empty torrent",
			torrent: &Torrent{
				Title: "Empty Album",
				Files: []FileLike{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, tt.name+".json")
			err := tt.torrent.Save(filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify file exists
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					t.Errorf("Save() file was not created: %s", filename)
				}
				// Verify file can be read back
				var loaded Torrent
				data, err := os.ReadFile(filename)
				if err != nil {
					t.Errorf("Save() failed to read file: %v", err)
					return
				}
				if err := json.Unmarshal(data, &loaded); err != nil {
					t.Errorf("Save() failed to unmarshal saved file: %v", err)
				}
			}
		})
	}
}
