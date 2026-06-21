package main

import (
	"context"
	"fmt"
	"os"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd := app.NewRootCommand(fullVersion())
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
