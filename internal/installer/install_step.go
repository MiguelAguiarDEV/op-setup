package installer

import (
	"context"
	"fmt"
	"time"
)

// InstallStep wraps an Installer to implement pipeline.Step and pipeline.RollbackStep.
// It skips installation if the tool is already detected.
//
// Timeout controls the per-installer deadline. If zero, no timeout is applied.
// The context is created internally in Run/Rollback with proper cancellation.
type InstallStep struct {
	Installer Installer
	Timeout   time.Duration
	installed bool
	skipped   bool
}

// ID returns a step identifier in the format "install-<installer-id>".
func (s *InstallStep) ID() string {
	return fmt.Sprintf("install-%s", s.Installer.ID())
}

// Run checks if the tool is already installed and skips if so.
// Otherwise, it runs the installer.
func (s *InstallStep) Run() error {
	ctx, cancel := s.contextWithTimeout()
	defer cancel()

	// Detect: skip if already installed.
	detected, err := s.Installer.Detect(ctx)
	if err != nil {
		return fmt.Errorf("detect %s: %w", s.Installer.Name(), err)
	}
	if detected {
		s.skipped = true
		return nil
	}

	// Install.
	if err := s.Installer.Install(ctx); err != nil {
		return err
	}
	s.installed = true
	return nil
}

// Rollback undoes the installation. Only rolls back if Install was actually called.
func (s *InstallStep) Rollback() error {
	if !s.installed {
		return nil
	}
	ctx, cancel := s.contextWithTimeout()
	defer cancel()
	return s.Installer.Rollback(ctx)
}

// Skipped returns true if the tool was already detected and installation was skipped.
func (s *InstallStep) Skipped() bool {
	return s.skipped
}

// Installed returns true if Install was actually executed.
func (s *InstallStep) Installed() bool {
	return s.installed
}

// contextWithTimeout creates a context with the configured timeout.
// If Timeout is zero, returns context.Background() with a no-op cancel.
func (s *InstallStep) contextWithTimeout() (context.Context, context.CancelFunc) {
	if s.Timeout > 0 {
		return context.WithTimeout(context.Background(), s.Timeout)
	}
	return context.Background(), func() {}
}
