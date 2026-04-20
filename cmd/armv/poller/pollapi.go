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

// PollApi drives respPoller to completion, displaying a progress bar while
// polling. The poll cycle is bounded by pollingTimeout and respects context
// cancellation at every wait point. Progress-bar render errors are treated
// as non-fatal because they do not affect the correctness of the underlying
// Azure operation.
func PollApi[T any](ctx context.Context, respPoller *runtime.Poller[T], outputPath string) error {
	ctx, cancel := context.WithTimeout(ctx, pollingTimeout)
	defer cancel()

	bar := progressBar()

	// Single reusable timer avoids allocating a new Timer on every iteration,
	// which would otherwise accumulate across a 30-minute polling window.
	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	barCount := 0
	for {
		select {
		case <-ctx.Done():
			_ = bar.Finish()
			return fmt.Errorf("polling timeout or cancelled: %w", ctx.Err())
		case <-timer.C:
		}
		timer.Reset(sleepDuration)

		barCount++
		_ = bar.Add(1) // progress bar render errors are non-fatal
		if barCount >= progressBarMax {
			bar.Reset()
			barCount = 0
		}

		w, err := respPoller.Poll(ctx)
		if err != nil {
			return fmt.Errorf("poll: %w", err)
		}

		if !respPoller.Done() {
			continue
		}

		_ = bar.Finish() // terminal render errors are non-fatal

		var respBody []byte
		if w != nil && w.Body != nil {
			respBody, err = io.ReadAll(w.Body)
			_ = w.Body.Close()
			if err != nil {
				return fmt.Errorf("read response body: %w", err)
			}
		}

		statusCode := 0
		status := ""
		if w != nil {
			statusCode = w.StatusCode
			status = w.Status
		}

		pollResp := NewPollerResponseData(respBody, statusCode, status)
		if err := pollResp.writeOutput(outputPath); err != nil {
			return err
		}
		return nil
	}
}
