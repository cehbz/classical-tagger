package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"github.com/cehbz/classical-tagger/internal/domain"
	"github.com/cehbz/classical-tagger/internal/filesystem"
	"github.com/cehbz/classical-tagger/internal/storage"
	"github.com/cehbz/classical-tagger/internal/validation"
)

// Example demonstrates the complete workflow:
// 1. Create an album from metadata
// 2. Validate the metadata
// 3. Validate the directory structure
// 4. Save to JSON
func main() {
	// Step 1: Create an album
	album := createSampleAlbum()
	
	// Step 2: Validate metadata
	fmt.Println("=== Metadata Validation ===")
	metadataValidator := validation.NewAlbumValidator()
	metadataIssues := metadataValidator.ValidateMetadata(album)
	
	if len(metadataIssues) == 0 {
		fmt.Println("✓ No metadata issues found")
	} else {
		fmt.Printf("Found %d metadata issues:\n", len(metadataIssues))
		for _, issue := range metadataIssues {
			fmt.Printf("  %s\n", issue)
		}
	}
	
	// Step 3: Validate directory structure
	fmt.Println("\n=== Directory Validation ===")
	dirValidator := filesystem.NewDirectoryValidator()
	
	// Validate folder name
	folderName := "RIAS Kammerchor - Christmas Motets (2013) - FLAC"
	folderIssues := dirValidator.ValidateFolderName(folderName, album)
	
	if len(folderIssues) == 0 {
		fmt.Println("✓ No folder naming issues found")
	} else {
		fmt.Printf("Found %d folder issues:\n", len(folderIssues))
		for _, issue := range folderIssues {
			fmt.Printf("  %s\n", issue)
		}
	}
	
	// Validate file paths
	samplePath := folderName + "/01 Frohlocket, Op. 79-1.flac"
	pathIssues := dirValidator.ValidatePath(samplePath)
	
	if len(pathIssues) == 0 {
		fmt.Println("✓ No path issues found")
	} else {
		fmt.Printf("Found %d path issues:\n", len(pathIssues))
		for _, issue := range pathIssues {
			fmt.Printf("  %s\n", issue)
		}
	}
	
	// Step 4: Save to JSON
	fmt.Println("\n=== JSON Serialization ===")
	repo := storage.NewRepository()
	jsonData, err := repo.SaveToJSON(album)
	if err != nil {
		log.Fatalf("Failed to save JSON: %v", err)
	}
	
	// Pretty print JSON
	var prettyJSON map[string]interface{}
	json.Unmarshal(jsonData, &prettyJSON)
	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Println(string(formatted))
	
	// Step 5: Load from JSON (round-trip)
	fmt.Println("\n=== JSON Deserialization ===")
	loadedAlbum, err := repo.LoadFromJSON(jsonData)
	if err != nil {
		log.Fatalf("Failed to load JSON: %v", err)
	}
	
	fmt.Printf("✓ Successfully loaded album: %s (%d)\n", loadedAlbum.Title(), loadedAlbum.OriginalYear())
	fmt.Printf("  Tracks: %d\n", len(loadedAlbum.Tracks()))
	
	// Summary
	fmt.Println("\n=== Summary ===")
	totalIssues := len(metadataIssues) + len(folderIssues) + len(pathIssues)
	if totalIssues == 0 {
		fmt.Println("✓ Album is fully compliant with all rules")
	} else {
		fmt.Printf("⚠ Found %d total validation issues\n", totalIssues)
	}
}

func createSampleAlbum() *domain.Album {
	// Create album
	album, _ := domain.NewAlbum("Noël ! Weihnachten ! Christmas!", 2013)
	
	// Add edition
	edition, _ := domain.NewEdition("harmonia mundi", 2013)
	edition = edition.WithCatalogNumber("HMC902170")
	album = album.WithEdition(edition)
	
	// Create artists
	mendelssohn, _ := domain.NewArtist("Felix Mendelssohn Bartholdy", domain.RoleComposer)
	ensemble, _ := domain.NewArtist("RIAS Kammerchor Berlin", domain.RoleEnsemble)
	conductor, _ := domain.NewArtist("Hans-Christoph Rademann", domain.RoleConductor)
	
	// Add track 1
	track1, _ := domain.NewTrack(
		1,
		1,
		"Frohlocket, ihr Völker auf Erden, Op. 79/1",
		[]domain.Artist{mendelssohn, ensemble, conductor},
	)
	track1 = track1.WithName("01 Frohlocket, Op. 79-1.flac")
	album.AddTrack(track1)
	
	// Add track 2
	brahms, _ := domain.NewArtist("Johannes Brahms", domain.RoleComposer)
	track2, _ := domain.NewTrack(
		1,
		2,
		"O Heiland, reiß die Himmel auf, Op. 74/2",
		[]domain.Artist{brahms, ensemble, conductor},
	)
	track2 = track2.WithName("02 O Heiland, Op. 74-2.flac")
	album.AddTrack(track2)
	
	// Add track 3
	poulenc, _ := domain.NewArtist("Francis Poulenc", domain.RoleComposer)
	track3, _ := domain.NewTrack(
		1,
		3,
		"O magnum mysterium",
		[]domain.Artist{poulenc, ensemble, conductor},
	)
	track3 = track3.WithName("03 O magnum mysterium.flac")
	album.AddTrack(track3)
	
	return album
}
