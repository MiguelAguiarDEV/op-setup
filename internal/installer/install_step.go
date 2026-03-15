package installer

import (
	"context"
	"fmt"
)

// InstallStep wraps an Installer to implement pipeline.Step and pipeline.RollbackStep.
// It skips installation if the tool is already detected.
//
// Ctx is stored as a struct field (rather than passed to Run/Rollback) because
// the pipeline.Step interface defines Run() error with no context parameter.
// InstallStep bridges this gap by holding the context for the underlying Installer.
type InstallStep struct {
	Installer Installer
	Ctx       context.Context
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
	ctx := s.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

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
	ctx := s.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
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
