package poller

import (
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// progressBar creates and returns a new progress bar with custom options.
//
// No parameters.
// Returns a pointer to progressbar.ProgressBar.
func progressBar() *progressbar.ProgressBar {

	bar := progressbar.NewOptions(progressBarMax,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription("[cyan][reset] Running Validation..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	return bar
}
