/*
MIT License

# Copyright (c) 2024 Aaron Saikovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
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
