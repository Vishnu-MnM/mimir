package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePaths(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Paths
		wantErr bool
	}{
		{
			name: "uses XDG vars when set",
			env: map[string]string{
				"HOME":            "/home/u",
				"XDG_CONFIG_HOME": "/xdg/config",
				"XDG_DATA_HOME":   "/xdg/data",
				"XDG_STATE_HOME":  "/xdg/state",
			},
			want: Paths{
				Config: "/xdg/config/mimir",
				Data:   "/xdg/data/mimir",
				State:  "/xdg/state/mimir",
			},
		},
		{
			name: "falls back to HOME defaults",
			env:  map[string]string{"HOME": "/home/u"},
			want: Paths{
				Config: "/home/u/.config/mimir",
				Data:   "/home/u/.local/share/mimir",
				State:  "/home/u/.local/state/mimir",
			},
		},
		{
			// The XDG spec says relative paths must be ignored.
			name: "ignores relative XDG paths",
			env:  map[string]string{"HOME": "/home/u", "XDG_CONFIG_HOME": "relative/dir"},
			want: Paths{
				Config: "/home/u/.config/mimir",
				Data:   "/home/u/.local/share/mimir",
				State:  "/home/u/.local/state/mimir",
			},
		},
		{
			name:    "errors when HOME is unset and no XDG fallback exists",
			env:     map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolvePaths(func(k string) string { return tt.env[k] })
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEnsureDirs(t *testing.T) {
	t.Run("creates missing directories with mode 0700", func(t *testing.T) {
		t.Parallel()
		p := testPaths(t.TempDir())

		if err := p.EnsureDirs(); err != nil {
			t.Fatalf("EnsureDirs() error = %v", err)
		}
		for _, dir := range dirsOf(p) {
			assertDirmode(t, dir, 0o700)
		}
	})

	t.Run("leaves existing directories and their modes alone", func(t *testing.T) {
		t.Parallel()
		p := testPaths(t.TempDir())
		for _, dir := range dirsOf(p) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			// Chmod explicitly: MkdirAll's mode is reduced by the umask.
			if err := os.Chmod(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}

		if err := p.EnsureDirs(); err != nil {
			t.Fatalf("EnsureDirs() err = %v", err)
		}
		for _, dir := range dirsOf(p) {
			assertDirmode(t, dir, 0o755)
		}
	})
}

func TestEnsureDirsErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) Paths
	}{
		{
			name:  "zero-value Paths",
			setup: func(_ *testing.T) Paths { return Paths{} },
		},
		{
			name: "relative path",
			setup: func(t *testing.T) Paths {
				p := testPaths(t.TempDir())
				p.State = "relative/state"
				return p
			},
		},
		{
			name: "file where a directory should be",
			setup: func(t *testing.T) Paths {
				p := testPaths(t.TempDir())
				if err := os.MkdirAll(filepath.Dir(p.Data), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p.Data, nil, 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := tt.setup(t)
			if err := p.EnsureDirs(); err == nil {
				t.Fatal("EnsureDirs() error = nil, want an error")
			}
		})
	}
}

// ---helpers---

func testPaths(root string) Paths {
	return Paths{
		Config: filepath.Join(root, "config", "mimir"),
		Data:   filepath.Join(root, "data", "mimir"),
		State:  filepath.Join(root, "state", "mimir"),
	}
}

func dirsOf(p Paths) []string {
	return []string{p.Config, p.Data, p.State}
}

func assertDirmode(t *testing.T, dir string, want fs.FileMode) {
	t.Helper()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat %s: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", dir)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode = %#o, want %#o", dir, got, want)
	}
}
