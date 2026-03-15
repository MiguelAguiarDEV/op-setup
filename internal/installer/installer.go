// Package installer provides tool installation, detection, and rollback.
package installer

import (
	"context"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Installer can detect, install, and rollback a tool.
type Installer interface {
	// ID returns the unique installer identifier.
	ID() model.InstallerID

	// Name returns the human-readable tool name.
	Name() string

	// Detect checks if the tool is already installed.
	// Returns true if installed, false otherwise.
	Detect(ctx context.Context) (bool, error)

	// Install performs the tool installation.
	Install(ctx context.Context) error

	// Rollback undoes the installation (best-effort).
	Rollback(ctx context.Context) error

	// Prerequisites returns binary names that must exist before install.
	// Empty slice means no prerequisites.
	Prerequisites() []string
}
