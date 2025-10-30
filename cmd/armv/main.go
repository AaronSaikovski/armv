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
	_ "embed"
	"os"

	"github.com/AaronSaikovski/armv/cmd/armv/app"
)

//ref: https://levelup.gitconnected.com/a-better-way-than-ldflags-to-add-a-build-version-to-your-go-binaries-2258ce419d2d

//go:generate bash get_version.sh
//go:embed version.txt
var version string

// main is the entry point of the program.
//
// It creates the root cobra command and executes it with a cancellable context.
// If an error occurs during command execution, the program exits with status code 1.
func main() {

	// Create a context with cancellation capability
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and execute the root command
	rootCmd := app.NewRootCommand(version)
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
