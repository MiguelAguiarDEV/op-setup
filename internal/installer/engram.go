package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// EngramInstaller installs engram via go install.
type EngramInstaller struct {
	Cmd        CommandRunner
	HomeDir    string
	RemoveFunc func(string) error
}

func (i *EngramInstaller) ID() model.InstallerID { return model.InstallerEngram }
func (i *EngramInstaller) Name() string          { return "Engram" }

func (i *EngramInstaller) Prerequisites() []string {
	return []string{"go"}
}

func (i *EngramInstaller) Detect(ctx context.Context) (bool, error) {
	_, err := i.Cmd.LookPath("engram")
	return err == nil, nil
}

func (i *EngramInstaller) Install(ctx context.Context) error {
	out, err := i.Cmd.Run(ctx, "go", "install",
		"github.com/Gentleman-Programming/engram/cmd/engram@latest")
	if err != nil {
		return &InstallFailedError{
			Installer: i.ID(),
			Reason:    err.Error(),
			Output:    string(out),
		}
	}
	return nil
}

func (i *EngramInstaller) Rollback(ctx context.Context) error {
	remove := i.RemoveFunc
	if remove == nil {
		remove = os.Remove
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(i.HomeDir, "go")
	}
	binPath := filepath.Join(gopath, "bin", "engram")
	if err := remove(binPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rollback %s: remove %s: %w", i.ID(), binPath, err)
	}
	return nil
}
