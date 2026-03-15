package installer

import (
	"errors"
	"strings"
	"testing"
)

func TestPrerequisiteMissingError(t *testing.T) {
	tests := []struct {
		name      string
		err       *PrerequisiteMissingError
		wantMsg   string
		wantIs    error
		wantNotIs error
	}{
		{
			name:      "basic",
			err:       &PrerequisiteMissingError{Installer: "opencode", Binary: "npm"},
			wantMsg:   `prerequisite "npm" not found for installer "opencode"`,
			wantIs:    ErrPrerequisiteMissing,
			wantNotIs: ErrInstallFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
			if !errors.Is(tt.err, tt.wantIs) {
				t.Errorf("Is(%v) = false, want true", tt.wantIs)
			}
			if errors.Is(tt.err, tt.wantNotIs) {
				t.Errorf("Is(%v) = true, want false", tt.wantNotIs)
			}
		})
	}
}

func TestInstallFailedError(t *testing.T) {
	tests := []struct {
		name    string
		err     *InstallFailedError
		wantSub string
	}{
		{
			name:    "with output",
			err:     &InstallFailedError{Installer: "engram", Reason: "exit 1", Output: "error log"},
			wantSub: "output: error log",
		},
		{
			name:    "without output",
			err:     &InstallFailedError{Installer: "engram", Reason: "exit 1"},
			wantSub: `install "engram" failed: exit 1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if !strings.Contains(msg, tt.wantSub) {
				t.Errorf("Error() = %q, want substring %q", msg, tt.wantSub)
			}
			if !errors.Is(tt.err, ErrInstallFailed) {
				t.Error("Is(ErrInstallFailed) = false, want true")
			}
			if errors.Is(tt.err, ErrPrerequisiteMissing) {
				t.Error("Is(ErrPrerequisiteMissing) = true, want false")
			}
		})
	}
}

func TestDuplicateInstallerError(t *testing.T) {
	err := &DuplicateInstallerError{ID: "opencode"}
	want := `installer already registered: "opencode"`
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrDuplicateInstaller) {
		t.Error("Is(ErrDuplicateInstaller) = false, want true")
	}
	if errors.Is(err, ErrInstallFailed) {
		t.Error("Is(ErrInstallFailed) = true, want false")
	}
}
