package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// PlaywrightInstaller installs Playwright's Chromium browser.
type PlaywrightInstaller struct {
	Cmd      CommandRunner
	HomeDir  string
	StatPath func(string) (os.FileInfo, error)
	GlobFunc func(string) ([]string, error)
}

func (i *PlaywrightInstaller) ID() model.InstallerID { return model.InstallerPlaywright }
func (i *PlaywrightInstaller) Name() string          { return "Playwright (Chromium)" }

func (i *PlaywrightInstaller) Prerequisites() []string {
	return []string{"npx"}
}

func (i *PlaywrightInstaller) Detect(ctx context.Context) (bool, error) {
	// Check if playwright chromium cache exists by globbing for chromium-*.
	glob := i.GlobFunc
	if glob == nil {
		glob = filepath.Glob
	}
	pattern := filepath.Join(i.chromiumCacheDir(), "chromium-*")
	matches, err := glob(pattern)
	if err != nil {
		return false, fmt.Errorf("detect %s: %w", i.ID(), err)
	}
	return len(matches) > 0, nil
}

func (i *PlaywrightInstaller) Install(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npx", "playwright", "install", "chromium")
	if err != nil {
		return &InstallFailedError{
			Installer: i.ID(),
			Reason:    err.Error(),
			Output:    string(out),
		}
	}
	return nil
}

func (i *PlaywrightInstaller) Rollback(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npx", "playwright", "uninstall", "chromium")
	if err != nil {
		return fmt.Errorf("rollback %s: %w\noutput: %s", i.ID(), err, string(out))
	}
	return nil
}

// chromiumCacheDir returns the OS-specific playwright cache directory.
func (i *PlaywrightInstaller) chromiumCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(i.HomeDir, "Library", "Caches", "ms-playwright")
	default: // linux
		return filepath.Join(i.HomeDir, ".cache", "ms-playwright")
	}
}
