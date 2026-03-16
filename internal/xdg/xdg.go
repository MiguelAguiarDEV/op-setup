// Package xdg provides XDG Base Directory helpers.
package xdg

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the XDG config base directory.
// If XDG_CONFIG_HOME is set and non-empty, it is returned as-is.
// Otherwise, falls back to filepath.Join(homeDir, ".config").
func ConfigDir(homeDir string) string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	return filepath.Join(homeDir, ".config")
}
