package installer

import (
	"errors"
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Sentinel errors for installer operations.
var (
	ErrPrerequisiteMissing = errors.New("prerequisite not found")
	ErrInstallFailed       = errors.New("installation failed")
	ErrDuplicateInstaller  = errors.New("installer already registered")
)

// PrerequisiteMissingError is returned when a required binary is not in PATH.
type PrerequisiteMissingError struct {
	Installer model.InstallerID
	Binary    string
}

func (e *PrerequisiteMissingError) Error() string {
	return fmt.Sprintf("prerequisite %q not found for installer %q", e.Binary, e.Installer)
}

func (e *PrerequisiteMissingError) Is(target error) bool {
	return target == ErrPrerequisiteMissing
}

// InstallFailedError is returned when a tool installation command fails.
type InstallFailedError struct {
	Installer model.InstallerID
	Reason    string
	Output    string
}

func (e *InstallFailedError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("install %q failed: %s\noutput: %s", e.Installer, e.Reason, e.Output)
	}
	return fmt.Sprintf("install %q failed: %s", e.Installer, e.Reason)
}

func (e *InstallFailedError) Is(target error) bool {
	return target == ErrInstallFailed
}

// DuplicateInstallerError is returned when registering an installer that already exists.
type DuplicateInstallerError struct {
	ID model.InstallerID
}

func (e *DuplicateInstallerError) Error() string {
	return fmt.Sprintf("installer already registered: %q", e.ID)
}

func (e *DuplicateInstallerError) Is(target error) bool {
	return target == ErrDuplicateInstaller
}
