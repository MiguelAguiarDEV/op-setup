package installer

import (
	"context"
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// ContextModeInstaller installs context-mode via npm.
type ContextModeInstaller struct {
	Cmd CommandRunner
}

func (i *ContextModeInstaller) ID() model.InstallerID { return model.InstallerContextMode }
func (i *ContextModeInstaller) Name() string          { return "Context Mode" }

func (i *ContextModeInstaller) Prerequisites() []string {
	return []string{"npm"}
}

func (i *ContextModeInstaller) Detect(ctx context.Context) (bool, error) {
	// Primary: check PATH.
	if _, err := i.Cmd.LookPath("context-mode"); err == nil {
		return true, nil
	}
	// Fallback: check npm global list.
	out, err := i.Cmd.Run(ctx, "npm", "list", "-g", "context-mode", "--depth=0")
	if err == nil && len(out) > 0 {
		return true, nil
	}
	return false, nil
}

func (i *ContextModeInstaller) Install(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npm", "install", "-g", "context-mode")
	if err != nil {
		return &InstallFailedError{
			Installer: i.ID(),
			Reason:    err.Error(),
			Output:    string(out),
		}
	}
	return nil
}

func (i *ContextModeInstaller) Rollback(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "npm", "uninstall", "-g", "context-mode")
	if err != nil {
		return fmt.Errorf("rollback %s: %w\noutput: %s", i.ID(), err, string(out))
	}
	return nil
}
