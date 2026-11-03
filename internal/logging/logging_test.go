package logging_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Vishnu-MnM/mimir/internal/logging"
)

func TestNew(t *testing.T) {
	t.Run("writes records at or above the level", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		logger := logging.New(&buf, slog.LevelInfo)

		logger.Info("task added", "id", "abc")

		got := buf.String()
		for _, want := range []string{"level=INFO", `msg="task added"`, "id=abc"} {
			if !strings.Contains(got, want) {
				t.Errorf("got %q is missing %q", got, want)
			}
		}
	})

	t.Run("drops records below the level", func(t *testing.T) {
		t.Parallel()
		var buf bytes.Buffer
		logger := logging.New(&buf, slog.LevelInfo)

		logger.Debug("noisy detail")

		if buf.Len() != 0 {
			t.Errorf("output = %q, want nothing", buf.String())
		}
	})
}

func TestOpenFile(t *testing.T) {
	t.Run("creates mimir.log with mode 0600", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		f, err := logging.OpenFile(dir)
		if err != nil {
			t.Fatalf("OpenFile() error = %v", err)
		}
		closeFile(t, f)

		info, err := os.Stat(filepath.Join(dir, "mimir.log"))
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("mode = %#o, want %#o", got, 0o600)
		}
	})

	t.Run("appends to an existing file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		writeViaOpenFile(t, dir, "first\n")
		writeViaOpenFile(t, dir, "second\n")

		got, err := os.ReadFile(filepath.Join(dir, "mimir.log"))
		if err != nil {
			t.Fatalf("readfile: %v", err)
		}
		if want := "first\nsecond\n"; string(got) != want {
			t.Errorf("content = %q, want %q", got, want)
		}
	})

	t.Run("errors when the directory is missing", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(t.TempDir(), "missing")

		f, err := logging.OpenFile(dir)
		if err == nil {
			closeFile(t, f)
			t.Fatalf("OpenFile() error = nil, want an error")
		}
	})
}

// ---helpers---
// writeViaOpenFile opens the log in dir, writes s, and closes it again,
// like one run of mimir would.

func writeViaOpenFile(t *testing.T, dir, s string) {
	t.Helper()
	f, err := logging.OpenFile(dir)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if _, err := f.WriteString(s); err != nil {
		t.Fatalf("write: %v", err)
	}
	closeFile(t, f)
}

func closeFile(t *testing.T, f *os.File) {
	t.Helper()
	if err := f.Close(); err != nil {
		t.Errorf("close %v", err)
	}
}
