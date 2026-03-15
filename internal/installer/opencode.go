package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// OpenCodeInstaller installs opencode-ai via npm.
type OpenCodeInstaller struct {
	Cmd      CommandRunner
	HomeDir  string
	StatPath func(string) (os.FileInfo, error)
}

func (i *OpenCodeInstaller) ID() model.InstallerID { return model.InstallerOpenCode }
func (i *OpenCodeInstaller) Name() string          { return "OpenCode" }

func (i *OpenCodeInstaller) Prerequisites() []string {
	return []string{"npm"}
}

func (i *OpenCodeInstaller) Detect(ctx context.Context) (bool, error) {
	// Primary: check PATH.
	if _, err := i.Cmd.LookPath("opencode"); err == nil {
		return true, nil
	}
	// Fallback: check ~/.opencode/bin/opencode.
	stat := i.StatPath
	if stat == nil {
		stat = os.Stat
	}
	binPath := filepath.Join(i.HomeDir, ".opencode", "bin", "opencode")
	if _, err := stat(binPath); err == nil {
		return true, nil
	}
	return false, nil
}

func (i *OpenCodeInstaller) Install(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npm", "install", "-g", "opencode-ai")
	if err != nil {
		return &InstallFailedError{
			Installer: i.ID(),
			Reason:    err.Error(),
			Output:    string(out),
		}
	}
	return nil
}

func (i *OpenCodeInstaller) Rollback(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npm", "uninstall", "-g", "opencode-ai")
	if err != nil {
		return fmt.Errorf("rollback %s: %w\noutput: %s", i.ID(), err, string(out))
	}
	return nil
}
