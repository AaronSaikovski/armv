package validation

import (
	"log/slog"
	"os"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
)
