package validation

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// MultiDiscFolderSorting checks multi-disc folder naming for proper sorting (rule 2.3.14.2)
// INFO level - suggests folder names that sort properly
func (r *Rules) MultiDiscFolderSorting(actualAlbum, _ *domain.Album) RuleResult {
	meta := RuleMetadata{
		ID:     "2.3.19",
		Name:   "Multi-disc folders should sort properly (CD1, CD2, not CD1, CD10, CD2)",
		Level:  domain.LevelInfo,
		Weight: 0.1,
	}

	var issues []domain.ValidationIssue

	if actualAlbum == nil || !actualAlbum.IsMultiDisc() {
		return RuleResult{Meta: meta, Issues: nil}
	}

	type folderDisc struct {
		folder string
		disc   int
	}

	seen := make(map[folderDisc]struct{})
	var tuples []folderDisc

	for _, track := range actualAlbum.Tracks {
		if track == nil || track.Name == "" {
			continue
		}

		// Convert to slash-separated, clean canonical path, extract first dir
		cleanPath := filepath.Clean(track.Name)
		components := strings.Split(cleanPath, string(filepath.Separator))
		if components[0] == "" {
			components = components[1:]
		}
		if len(components) == 0 {
			continue // no directory component
		}
		folder := components[0]

		if _, exists := seen[folderDisc{folder: folder, disc: track.Disc}]; exists {
			continue
		}
		seen[folderDisc{folder: folder, disc: track.Disc}] = struct{}{}
		tuples = append(tuples, folderDisc{folder: folder, disc: track.Disc})
	}

	if len(tuples) <= 1 {
		return RuleResult{Meta: meta, Issues: nil}
	}

	sort.Slice(tuples, func(i, j int) bool {
		if tuples[i].folder == tuples[j].folder {
			return tuples[i].disc < tuples[j].disc
		}
		return tuples[i].folder < tuples[j].folder
	})

	prev := tuples[0]
	for i := 1; i < len(tuples); i++ {
		current := tuples[i]
		if current.disc <= prev.disc {
			issues = append(issues, domain.ValidationIssue{
				Level: domain.LevelInfo,
				Track: 0,
				Rule:  meta.ID,
				Message: fmt.Sprintf("Folder '%s' (disc %d) should come after folder '%s' (disc %d) to maintain disc order",
					current.folder, current.disc, prev.folder, prev.disc),
			})
		}
		prev = current
	}

	return RuleResult{Meta: meta, Issues: issues}
}
