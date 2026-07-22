package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AaronSaikovski/armv/cmd/armv/app"
)

// Build metadata injected at link time via:
//
//	-ldflags "-X main.version=... -X main.commit=... -X main.date=..."
//
// Release (Taskfile / goreleaser) pipelines populate all three.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// fullVersion returns the full version string shown by --version.
func fullVersion() string {
	return fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}

func main() {
	// Cancel the context on Ctrl-C / SIGTERM so the poll loop unwinds cleanly
	// (finishes the progress bar and returns a cancellation error) instead of
	// the process being hard-killed mid-operation.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd := app.NewRootCommand(fullVersion())
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		// Cobra is configured with SilenceErrors, so surface the error here.
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
