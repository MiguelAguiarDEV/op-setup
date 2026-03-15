package installer

import (
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestNewDefaultRegistry(t *testing.T) {
	r, err := NewDefaultRegistry("/home/test")
	if err != nil {
		t.Fatalf("NewDefaultRegistry() error = %v", err)
	}

	wantIDs := []model.InstallerID{
		model.InstallerContextMode,
		model.InstallerEngram,
		model.InstallerOpenCode,
		model.InstallerPlaywright,
	}
	all := r.All()

	if len(all) != len(wantIDs) {
		t.Fatalf("All() len = %d, want %d", len(all), len(wantIDs))
	}

	for i, want := range wantIDs {
		if all[i].ID() != want {
			t.Errorf("All()[%d].ID() = %q, want %q", i, all[i].ID(), want)
		}
	}
}

func TestCompileTimeInterfaceChecks(t *testing.T) {
	// Compile-time checks in factory.go verify:
	// - All 4 installers implement Installer
	// - InstallStep implements pipeline.RollbackStep
	// This test confirms the file compiles correctly.
	t.Log("compile-time interface checks passed")
}
