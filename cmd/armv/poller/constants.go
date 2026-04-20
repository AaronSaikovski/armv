package poller

import "time"

const (
	progressBarMax = 100
	// Azure long-running operations typically take minutes; poll every 2s.
	sleepDuration  = 2 * time.Second
	pollingTimeout = 30 * time.Minute

	// HTTP status codes returned by the validate-move API.
	//   204 — validation succeeded, no issues found
	//   409 — validation failed with conflicts; JSON error body describes them
	StatusMoveOK      = 204
	StatusMoveFailure = 409
)
