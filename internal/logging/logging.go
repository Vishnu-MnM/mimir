// Package logging builds mimir's structured logger. Logs go to a file,
// never the terminal, because the TUI owns stdout and stderr.
package logging

import (
	"io"
	"log/slog"
)

// New returns a logger that writes text records at or above level to w.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}
