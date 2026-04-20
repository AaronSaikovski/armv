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

// PollApi drives respPoller to completion, writing a Markdown report to
// outputPath and returning the parsed ValidationReport. The poll cycle is
// bounded by pollingTimeout and respects context cancellation at every wait
// point. Progress-bar render errors are treated as non-fatal because they do
// not affect the correctness of the underlying Azure operation.
func PollApi[T any](
	ctx context.Context,
	respPoller *runtime.Poller[T],
	outputPath string,
	reportCtx ReportContext,
) (ValidationReport, error) {
	ctx, cancel := context.WithTimeout(ctx, pollingTimeout)
	defer cancel()

	bar := progressBar()

	// Reusable timer avoids allocating a new Timer per iteration across
	// the 30-minute polling window.
	timer := time.NewTimer(sleepDuration)
	defer timer.Stop()

	barCount := 0
	for {
		select {
		case <-ctx.Done():
			_ = bar.Finish()
			return ValidationReport{}, fmt.Errorf("polling timeout or cancelled: %w", ctx.Err())
		case <-timer.C:
		}
		timer.Reset(sleepDuration)

		barCount++
		_ = bar.Add(1) // progress-bar render errors are non-fatal
		if barCount >= progressBarMax {
			bar.Reset()
			barCount = 0
		}

		w, err := respPoller.Poll(ctx)
		if err != nil {
			return ValidationReport{}, fmt.Errorf("poll: %w", err)
		}

		if !respPoller.Done() {
			continue
		}

		_ = bar.Finish()

		var respBody []byte
		if w != nil && w.Body != nil {
			respBody, err = io.ReadAll(w.Body)
			_ = w.Body.Close()
			if err != nil {
				return ValidationReport{}, fmt.Errorf("read response body: %w", err)
			}
		}

		statusCode := 0
		status := ""
		if w != nil {
			statusCode = w.StatusCode
			status = w.Status
		}

		pollResp := NewPollerResponseData(respBody, statusCode, status)
		return pollResp.writeOutput(outputPath, reportCtx)
	}
}
