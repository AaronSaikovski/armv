package poller

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// TickFn is invoked once per polling iteration (roughly every sleepDuration) with
// the elapsed wall-clock time since polling began. It lets non-interactive callers
// emit progress updates without PollAndCollect having to know about MCP, logging,
// or any other transport. Pass nil to disable.
type TickFn func(elapsed time.Duration)

// PollAndCollect polls the long-running operation without touching stdout or writing
// files, returning the raw response once complete. Used by non-interactive callers
// (e.g. the MCP server) that must not emit anything on stdout. If onTick is non-nil,
// it is called on each poll iteration so callers can surface progress.
func PollAndCollect[T any](ctx context.Context, respPoller *runtime.Poller[T], onTick TickFn) (*PollerResponseData, error) {
	ctx, cancel := context.WithTimeout(ctx, pollingTimeout)
	defer cancel()

	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("polling timeout or cancelled: %w", ctx.Err())
		default:
		}

		time.Sleep(sleepDuration)

		if onTick != nil {
			onTick(time.Since(start))
		}

		w, err := respPoller.Poll(ctx)
		if err != nil {
			return nil, err
		}

		if respPoller.Done() {
			respBody, err := io.ReadAll(w.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			resp := NewPollerResponseData(respBody, w.StatusCode, w.Status)
			return &resp, nil
		}
	}
}

// ResourceMoveOK reports whether the HTTP status indicates a successful (empty-body) validation.
func ResourceMoveOK(statusCode int) bool {
	return statusCode == StatusMoveOK
}
