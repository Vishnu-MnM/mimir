// Package logging builds mimir's structured logger. Logs go to a file,
// never the terminal, because the TUI owns stdout and stderr.
package logging

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	// fileName is the log file inside the state directory
	fileName = "mimir.log"
	// fileMode keeps logs private: they may contain task titles.
	fileMode fs.FileMode = 0o600
)

// New returns a logger that writes text records at or above level to w.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}

// OpenFile opens dir/mimir.log for appending, creating it with mode 0600
// if needed. It does not create dir; call config.Paths.EnsureDirs first.
// The caller must close the returned file.
func OpenFile(dir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dir, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	return f, nil
}
