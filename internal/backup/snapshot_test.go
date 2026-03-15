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

func TestNewSnapshotter_DefaultNow(t *testing.T) {
	s := NewSnapshotter()
	if s.Now == nil {
		t.Fatal("Now should not be nil")
	}
	// Verify it returns a reasonable time (not zero).
	now := s.Now()
	if now.IsZero() {
		t.Fatal("Now() should return non-zero time")
	}
}

func TestSnapshotter_Create_UnreadableSource(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	// Create a file then make it unreadable.
	file := filepath.Join(homeDir, "secret.json")
	os.WriteFile(file, []byte("data"), 0o000)
	defer os.Chmod(file, 0o644)

	s := &Snapshotter{Now: fixedTime}
	_, err := s.Create(snapshotDir, []string{file})
	if err == nil {
		t.Fatal("expected error for unreadable source file")
	}
}

func TestReadManifest_MissingFile(t *testing.T) {
	_, err := ReadManifest("/nonexistent/manifest.json")
	if err == nil {
		t.Fatal("expected error for missing manifest file")
	}
}

func TestReadManifest_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := ReadManifest(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteManifest_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "deep", "manifest.json")

	m := Manifest{ID: "test", Entries: []ManifestEntry{}}
	if err := WriteManifest(path, m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was written.
	got, err := ReadManifest(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.ID != "test" {
		t.Fatalf("ID = %q, want %q", got.ID, "test")
	}
}

func TestWriteManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	original := Manifest{
		ID:      "20260315-120000",
		RootDir: dir,
		Entries: []ManifestEntry{
			{OriginalPath: "/home/user/.config/file.json", SnapshotPath: "/backup/0_file.json", Existed: true, Mode: 0o600},
			{OriginalPath: "/home/user/.config/new.json", Existed: false},
		},
	}

	if err := WriteManifest(path, original); err != nil {
		t.Fatal(err)
	}

	got, err := ReadManifest(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != original.ID {
		t.Fatalf("ID = %q, want %q", got.ID, original.ID)
	}
	if len(got.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(got.Entries))
	}
	if got.Entries[0].Mode != 0o600 {
		t.Fatalf("Mode = %o, want %o", got.Entries[0].Mode, 0o600)
	}
	if got.Entries[1].Existed {
		t.Fatal("second entry should not exist")
	}
}
