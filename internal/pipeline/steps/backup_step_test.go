package steps

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
)

func fixedTime() time.Time {
	return time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
}

func TestBackupStep_CreatesBackup(t *testing.T) {
	homeDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	// Create a config file.
	cfgFile := filepath.Join(homeDir, "settings.json")
	os.WriteFile(cfgFile, []byte(`{"key": "value"}`), 0o644)

	snap := &backup.Snapshotter{Now: fixedTime}
	step := NewBackupStep(snap, []string{cfgFile}, backupDir)

	if step.ID() != "backup-configs" {
		t.Fatalf("ID() = %q, want %q", step.ID(), "backup-configs")
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	m := step.Manifest()
	if m == nil {
		t.Fatal("manifest should not be nil after Run")
	}
	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	if !m.Entries[0].Existed {
		t.Fatal("entry should have Existed=true")
	}
}

func TestBackupStep_NonExistentFile(t *testing.T) {
	backupDir := filepath.Join(t.TempDir(), "backup")

	snap := &backup.Snapshotter{Now: fixedTime}
	step := NewBackupStep(snap, []string{"/nonexistent/file.json"}, backupDir)

	if err := step.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	m := step.Manifest()
	if m.Entries[0].Existed {
		t.Fatal("entry should have Existed=false")
	}
}

func TestBackupStep_ManifestNilBeforeRun(t *testing.T) {
	snap := backup.NewSnapshotter()
	step := NewBackupStep(snap, nil, "")

	if step.Manifest() != nil {
		t.Fatal("manifest should be nil before Run")
	}
}
