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

// Package poller provides functionality for polling Azure long-running operations
// and displaying progress to the user with timeout protection.
package poller

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// PollApi polls the API and displays the response.
//
// It takes a context.Context, a *runtime.Poller[T], and a string as parameters.
// The context is used for cancellation and timeout control.
// The respPoller is used to handle the polling.
// The outputPath is the path where the output is written.
// It returns an error if any occurred during the polling process.
func PollApi[T any](ctx context.Context, respPoller *runtime.Poller[T], outputPath string) error {

	// Add timeout protection to prevent infinite polling
	ctx, cancel := context.WithTimeout(ctx, pollingTimeout)
	defer cancel()

	//progress bar
	bar := progressBar()

	barCount := 0
	for {
		// Check if context has been cancelled or timed out
		select {
		case <-ctx.Done():
			_ = bar.Finish()
			return fmt.Errorf("polling timeout or cancelled: %w", ctx.Err())
		default:
			// Continue with polling
		}

		barCount++
		if err := bar.Add(1); err != nil {
			return err
		}

		time.Sleep(sleepDuration)

		if barCount >= progressBarMax {
			bar.Reset()
			barCount = 0
		}

		w, err := respPoller.Poll(ctx)
		if err != nil {
			return err
		}

		if respPoller.Done() {

			if err := bar.Finish(); err != nil {
				return err
			}

			// create new PollerResponseData
			respBody, err := io.ReadAll(w.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			pollResp := NewPollerResponseData(respBody, w.StatusCode, w.Status)

			// Write output to file
			if err := pollResp.writeOutput(outputPath); err != nil {
				return err
			}

			return nil

		}

	}
}
