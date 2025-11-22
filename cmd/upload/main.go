// cmd/upload/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cehbz/classical-tagger/internal/cache"
	"github.com/cehbz/classical-tagger/internal/uploader"
)

func main() {
	// Define flags
	var (
		torrentDir   = flag.String("dir", "", "Directory containing tagged FLAC files (required)")
		torrentID    = flag.Int("torrent", 0, "ID of torrent to trump (required)")
		apiKey       = flag.String("api-key", "", "Redacted API key (or set REDACTED_API_KEY env)")
		trumpReason  = flag.String("reason", "", "Custom trump reason (optional, auto-generated if not provided)")
		dryRun       = flag.Bool("dry-run", false, "Perform dry run without uploading")
		clearCache   = flag.Bool("clear-cache", false, "Clear metadata cache before running")
		verbose      = flag.Bool("verbose", false, "Enable verbose output")
		help         = flag.Bool("help", false, "Show help message")
	)
	
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Classical Music Torrent Uploader

This tool uploads a properly tagged and validated classical music torrent to Redacted,
typically to trump an existing torrent with incorrect tags or filenames.

Usage: %s [options]

Options:
`, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Environment Variables:
  REDACTED_API_KEY    Redacted API key (alternative to --api-key flag)
  XDG_CACHE_HOME      Cache directory (defaults to ~/.cache)

Examples:
  # Trump torrent 123456 with files in ./tagged_album/
  %s --dir ./tagged_album --torrent 123456

  # Dry run to see what would be uploaded
  %s --dir ./tagged_album --torrent 123456 --dry-run --verbose

  # Upload with custom trump reason
  %s --dir ./tagged_album --torrent 123456 --reason "Fixed composer names and work groupings"

  # Clear cache and re-fetch metadata
  %s --dir ./tagged_album --torrent 123456 --clear-cache

Workflow:
  1. Run 'extract' to fetch metadata from sources
  2. Run 'validate' to check for compliance issues
  3. Run 'tag' to fix tags and filenames
  4. Run 'upload' to trump the original torrent

The uploader will:
  - Fetch existing torrent and group metadata from Redacted
  - Validate artist consistency between Redacted and your tags
  - Preserve site metadata (tags, description, etc.)
  - Create a new .torrent file
  - Upload with appropriate trump reason

Cache:
  Metadata is cached for 24 hours to avoid repeated API calls.
  Use --clear-cache to force fresh fetches.

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}
	
	flag.Parse()
	
	// Show help if requested
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	
	// Validate required arguments
	if *torrentDir == "" {
		fmt.Fprintf(os.Stderr, "Error: --dir is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	if *torrentID == 0 {
		fmt.Fprintf(os.Stderr, "Error: --torrent is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// Get API key from flag or environment
	if *apiKey == "" {
		*apiKey = os.Getenv("REDACTED_API_KEY")
	}
	
	if *apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: API key required (use --api-key or set REDACTED_API_KEY)\n\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// Resolve torrent directory to absolute path
	absDir, err := filepath.Abs(*torrentDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving directory path: %v\n", err)
		os.Exit(1)
	}
	
	// Check directory exists
	if info, err := os.Stat(absDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: directory %s does not exist\n", absDir)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", absDir)
		os.Exit(1)
	}

	// Create upload command
	cmd := uploader.NewUploadCommand(*apiKey, absDir, *torrentID)
	
	// Configure options
	if *trumpReason != "" {
		cmd.TrumpReason = *trumpReason
	}
	cmd.DryRun = *dryRun
	cmd.Verbose = *verbose

	
	// Clear cache if requested
	if *clearCache {
		if *verbose {
			fmt.Println("Clearing cache...")
		}

		c := cache.NewCache(24 * time.Hour)
		if err := c.Clear("redacted-uploader"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clear cache: %v\n", err)
		}
	}
	
	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, cancelling upload...")
		cancel()
	}()
	
	// Execute upload
	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Upload failed: %v\n", err)
		os.Exit(1)
	}
	
	if *dryRun {
		fmt.Println("\nDry run completed successfully. No changes were made.")
	} else {
		fmt.Println("\nUpload completed successfully!")
	}
}