package installer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// mockInstaller is a configurable Installer for InstallStep tests.
type mockInstaller struct {
	id             model.InstallerID
	name           string
	detectResult   bool
	detectErr      error
	installErr     error
	rollbackErr    error
	prerequisites  []string
	installCalled  bool
	rollbackCalled bool
}

func (m *mockInstaller) ID() model.InstallerID { return m.id }
func (m *mockInstaller) Name() string          { return m.name }
func (m *mockInstaller) Detect(_ context.Context) (bool, error) {
	return m.detectResult, m.detectErr
}
func (m *mockInstaller) Install(_ context.Context) error {
	m.installCalled = true
	return m.installErr
}
func (m *mockInstaller) Rollback(_ context.Context) error {
	m.rollbackCalled = true
	return m.rollbackErr
}
func (m *mockInstaller) Prerequisites() []string { return m.prerequisites }

func TestInstallStep_ID(t *testing.T) {
	step := &InstallStep{
		Installer: &mockInstaller{id: model.InstallerOpenCode},
	}
	if got := step.ID(); got != "install-opencode" {
		t.Errorf("ID() = %q, want %q", got, "install-opencode")
	}
}

func TestInstallStep_Run(t *testing.T) {
	tests := []struct {
		name          string
		detectResult  bool
		detectErr     error
		installErr    error
		wantErr       bool
		wantSkipped   bool
		wantInstalled bool
	}{
		{
			name:          "already installed - skip",
			detectResult:  true,
			wantSkipped:   true,
			wantInstalled: false,
		},
		{
			name:          "not installed - install success",
			detectResult:  false,
			wantSkipped:   false,
			wantInstalled: true,
		},
		{
			name:         "not installed - install fails",
			detectResult: false,
			installErr:   &InstallFailedError{Installer: "test", Reason: "boom"},
			wantErr:      true,
		},
		{
			name:      "detect fails",
			detectErr: errors.New("detect error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &mockInstaller{
				id:           "test",
				name:         "Test",
				detectResult: tt.detectResult,
				detectErr:    tt.detectErr,
				installErr:   tt.installErr,
			}
			step := &InstallStep{
				Installer: inst,
			}

			err := step.Run()

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if step.Skipped() != tt.wantSkipped {
				t.Errorf("Skipped() = %v, want %v", step.Skipped(), tt.wantSkipped)
			}
			if step.Installed() != tt.wantInstalled {
				t.Errorf("Installed() = %v, want %v", step.Installed(), tt.wantInstalled)
			}
		})
	}
}

func TestInstallStep_Run_ZeroTimeout(t *testing.T) {
	inst := &mockInstaller{
		id:           "test",
		name:         "Test",
		detectResult: true,
	}
	step := &InstallStep{
		Installer: inst,
		// Timeout is zero — should use context.Background() (no deadline).
	}
	if err := step.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if !step.Skipped() {
		t.Error("Skipped() = false, want true")
	}
}

func TestInstallStep_Run_WithTimeout(t *testing.T) {
	inst := &mockInstaller{
		id:           "test",
		name:         "Test",
		detectResult: true,
	}
	step := &InstallStep{
		Installer: inst,
		Timeout:   5 * time.Minute,
	}
	if err := step.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if !step.Skipped() {
		t.Error("Skipped() = false, want true")
	}
}

func TestInstallStep_Rollback(t *testing.T) {
	tests := []struct {
		name               string
		installed          bool
		rollbackErr        error
		wantErr            bool
		wantRollbackCalled bool
	}{
		{
			name:               "rollback after install",
			installed:          true,
			wantRollbackCalled: true,
		},
		{
			name:               "no rollback when skipped",
			installed:          false,
			wantRollbackCalled: false,
		},
		{
			name:               "rollback error propagated",
			installed:          true,
			rollbackErr:        errors.New("rollback failed"),
			wantErr:            true,
			wantRollbackCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &mockInstaller{
				id:          "test",
				name:        "Test",
				rollbackErr: tt.rollbackErr,
			}
			step := &InstallStep{
				Installer: inst,
				installed: tt.installed,
			}

			err := step.Rollback()

			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
			if inst.rollbackCalled != tt.wantRollbackCalled {
				t.Errorf("rollbackCalled = %v, want %v", inst.rollbackCalled, tt.wantRollbackCalled)
			}
		})
	}
}

func TestInstallStep_Rollback_NilContext(t *testing.T) {
	inst := &mockInstaller{id: "test", name: "Test"}
	step := &InstallStep{
		Installer: inst,
		installed: true,
		// Ctx is nil
	}
	if err := step.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v", err)
	}
	if !inst.rollbackCalled {
		t.Error("rollbackCalled = false, want true")
	}
}

func TestInstallStep_ImplementsRollbackStep(t *testing.T) {
	// Compile-time interface check is in install_step_check_test.go (external package).
	// This test verifies the methods exist structurally.
	step := &InstallStep{
		Installer: &mockInstaller{id: "test", name: "Test"},
	}
	_ = step.ID()
	_ = step.Run()
	_ = step.Rollback()
}
