// Package config resolves mimir's per user directories using XDG base directory spec,
// loads, validates and applies the defaults to TOML config file.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

const appName = "mimir"

// Paths hold the per-user directories mimir needs to read and write to.
type Paths struct {
	Config string // config.toml, keybindings, user themes
	Data   string // SQLite Database
	State  string // logs, undo history, persisted pane widths.
}

// ErrNoHome is returned when a directory has no usable XDG variable and
// and HOME can't be used as a fallback.
var ErrNoHome = errors.New("home is not set to an absolute path")

// ResolvePaths works out mimir's directories using the XDG base directory spec.
// getenv is passed in (usually os.GetEnv) so tests never touch the real environment.
func ResolvePaths(getenv func(string) string) (Paths, error) {
	home := getenv("HOME")

	config, err := xdgDir(getenv, home, "XDG_CONFIG_HOME", ".config")
	if err != nil {
		return Paths{}, err
	}
	data, err := xdgDir(getenv, home, "XDG_DATA_HOME", ".local/share")
	if err != nil {
		return Paths{}, err
	}
	state, err := xdgDir(getenv, home, "XDG_STATE_HOME", ".local/state")
	if err != nil {
		return Paths{}, err
	}

	return Paths{Config: config, Data: data, State: state}, nil
}

// xdgDir returns $<envVar>/mimir if envVar holds an absolute path.
// Otherwise it falls back to $HOME/<fallback>/mimir.
func xdgDir(getenv func(string) string, home, envVar, fallback string) (string, error) {
	// The spec says relative values must be ignored. IsAbs("") is false,
	// so this one check covers both "unset" and "relative"
	if dir := getenv(envVar); filepath.IsAbs(dir) {
		return filepath.Join(dir, appName), nil
	}
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("resolving %s, %w", envVar, ErrNoHome)
	}
	return filepath.Join(home, fallback, appName), nil
}
