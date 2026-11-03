package logging_test

import (
	"bytes"
	"log/slog"
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
