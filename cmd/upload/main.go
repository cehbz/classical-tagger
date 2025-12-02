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

	"github.com/cehbz/classical-tagger/internal/cache"
	"github.com/cehbz/classical-tagger/internal/config"
	"github.com/cehbz/classical-tagger/internal/uploader"
)

func main() {
	// Define flags
	var (
		torrentDir  = flag.String("dir", "", "Directory containing tagged FLAC files (required)")
		torrentID   = flag.Int("torrent", 0, "ID of torrent to trump (required)")
		apiKey      = flag.String("api-key", "", "Redacted API key (optional, will be loaded from config file if not provided)")
		trumpReason = flag.String("reason", "", "Custom trump reason (optional, auto-generated if not provided)")
		dryRun      = flag.Bool("dry-run", false, "Perform dry run without uploading")
		clearCache  = flag.Bool("clear-cache", false, "Clear metadata cache before running")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		help        = flag.Bool("help", false, "Show help message")
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
		Configuration:
		  Config file location: %s
		  
		  The config file must contain your Redacted API key:
			redacted:
			  api_key: "your-key-here"
		  
		  Use --api-key flag to override config file.
		  
		  XDG_CACHE_HOME can be set to override cache directory (defaults to ~/.cache)
		`, config.GetConfigPathForDisplay())
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

	// Get API key from flag or config file
	if *apiKey == "" {
		var err error
		*apiKey, err = config.LoadRedactedAPIKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading API key from config: %v\n\n", err)
			fmt.Fprintf(os.Stderr, "Either use --api-key flag or configure redacted.api_key in:\n")
			fmt.Fprintf(os.Stderr, "  %s\n\n", config.GetConfigPathForDisplay())
			flag.Usage()
			os.Exit(1)
		}
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

		c := cache.NewCache(0)
		if err := c.Clear("redacted"); err != nil {
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
