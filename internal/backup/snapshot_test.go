package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fixedTime() time.Time {
	return time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
}

func TestSnapshotter_Create_TwoFiles(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	// Create two source files.
	file1 := filepath.Join(homeDir, "settings.json")
	file2 := filepath.Join(homeDir, "config.toml")
	os.WriteFile(file1, []byte(`{"key": "value1"}`), 0o644)
	os.WriteFile(file2, []byte(`key = "value2"`), 0o644)

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{file1, file2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manifest.ID != "20260315-120000" {
		t.Fatalf("ID = %q, want %q", manifest.ID, "20260315-120000")
	}
	if len(manifest.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(manifest.Entries))
	}

	// Verify first entry.
	e1 := manifest.Entries[0]
	if e1.OriginalPath != file1 {
		t.Fatalf("entry[0].OriginalPath = %q, want %q", e1.OriginalPath, file1)
	}
	if !e1.Existed {
		t.Fatal("entry[0].Existed should be true")
	}
	if e1.SnapshotPath == "" {
		t.Fatal("entry[0].SnapshotPath should not be empty")
	}

	// Verify backup file content.
	data, err := os.ReadFile(e1.SnapshotPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != `{"key": "value1"}` {
		t.Fatalf("backup content = %q", data)
	}
}

func TestSnapshotter_Create_NonExistentFile(t *testing.T) {
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{"/nonexistent/file.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(manifest.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(manifest.Entries))
	}

	e := manifest.Entries[0]
	if e.Existed {
		t.Fatal("Existed should be false for non-existent file")
	}
	if e.SnapshotPath != "" {
		t.Fatalf("SnapshotPath should be empty, got %q", e.SnapshotPath)
	}
}

func TestSnapshotter_Create_ManifestWritten(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	file1 := filepath.Join(homeDir, "settings.json")
	os.WriteFile(file1, []byte("{}"), 0o644)

	s := &Snapshotter{Now: fixedTime}
	_, err := s.Create(snapshotDir, []string{file1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify manifest file exists and is readable.
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	m, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if m.ID != "20260315-120000" {
		t.Fatalf("manifest ID = %q", m.ID)
	}
	if len(m.Entries) != 1 {
		t.Fatalf("manifest entries = %d", len(m.Entries))
	}
}

func TestSnapshotter_Create_PreservesPermissions(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	file1 := filepath.Join(homeDir, "secret.json")
	os.WriteFile(file1, []byte("secret"), 0o600)

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{file1})
	if err != nil {
		t.Fatal(err)
	}

	e := manifest.Entries[0]
	if e.Mode != 0o600 {
		t.Fatalf("Mode = %o, want %o", e.Mode, 0o600)
	}

	// Verify backup file has same permissions.
	info, _ := os.Stat(e.SnapshotPath)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("backup perm = %o, want %o", info.Mode().Perm(), 0o600)
	}
}

func TestSnapshotter_Create_EmptyPaths(t *testing.T) {
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(manifest.Entries))
	}
}

func TestSnapshotter_Create_MixedExistence(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	existing := filepath.Join(homeDir, "exists.json")
	os.WriteFile(existing, []byte("data"), 0o644)
	missing := filepath.Join(homeDir, "missing.json")

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{existing, missing})
	if err != nil {
		t.Fatal(err)
	}

	if len(manifest.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(manifest.Entries))
	}
	if !manifest.Entries[0].Existed {
		t.Fatal("first entry should exist")
	}
	if manifest.Entries[1].Existed {
		t.Fatal("second entry should not exist")
	}
}
