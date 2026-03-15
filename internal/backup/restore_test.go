package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreService_RestoresModifiedFile(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	// Create original file.
	original := filepath.Join(homeDir, "settings.json")
	os.WriteFile(original, []byte("original content"), 0o644)

	// Create backup.
	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{original})
	if err != nil {
		t.Fatal(err)
	}

	// Modify the original file (simulating what the installer would do).
	os.WriteFile(original, []byte("modified content"), 0o644)

	// Restore.
	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	// Verify original content restored.
	data, _ := os.ReadFile(original)
	if string(data) != "original content" {
		t.Fatalf("content = %q, want %q", data, "original content")
	}
}

func TestRestoreService_RemovesCreatedFile(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	// File doesn't exist yet.
	newFile := filepath.Join(homeDir, "new.json")

	// Create backup (file doesn't exist, so Existed=false).
	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{newFile})
	if err != nil {
		t.Fatal(err)
	}

	// Simulate installer creating the file.
	os.WriteFile(newFile, []byte("new content"), 0o644)

	// Restore — should remove the file.
	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	// Verify file was removed.
	if _, err := os.Stat(newFile); !os.IsNotExist(err) {
		t.Fatal("file should have been removed by restore")
	}
}

func TestRestoreService_RemoveNonExistent_NoError(t *testing.T) {
	// Restore with a file that was never created — should not error.
	manifest := Manifest{
		Entries: []ManifestEntry{
			{
				OriginalPath: "/tmp/never-existed-" + t.Name(),
				Existed:      false,
			},
		},
	}

	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatalf("restore should not error for non-existent file: %v", err)
	}
}

func TestRestoreService_MissingSnapshot_ReturnsError(t *testing.T) {
	manifest := Manifest{
		Entries: []ManifestEntry{
			{
				OriginalPath: "/tmp/test",
				SnapshotPath: "/nonexistent/backup/file",
				Existed:      true,
				Mode:         0o644,
			},
		},
	}

	rs := NewRestoreService()
	err := rs.Restore(manifest)
	if err == nil {
		t.Fatal("expected error for missing snapshot file")
	}
}

func TestRestoreService_EmptyManifest(t *testing.T) {
	manifest := Manifest{Entries: []ManifestEntry{}}

	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRestoreService_PreservesPermissions(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	original := filepath.Join(homeDir, "secret.json")
	os.WriteFile(original, []byte("secret"), 0o600)

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{original})
	if err != nil {
		t.Fatal(err)
	}

	// Modify file with different permissions.
	os.WriteFile(original, []byte("changed"), 0o644)

	// Restore.
	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatal(err)
	}

	info, _ := os.Stat(original)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %o, want %o", info.Mode().Perm(), 0o600)
	}
}

func TestRestoreService_MultipleEntries(t *testing.T) {
	homeDir := t.TempDir()
	snapshotDir := filepath.Join(t.TempDir(), "backup")

	file1 := filepath.Join(homeDir, "a.json")
	file2 := filepath.Join(homeDir, "b.json")
	file3 := filepath.Join(homeDir, "c.json") // won't exist initially

	os.WriteFile(file1, []byte("a-original"), 0o644)
	os.WriteFile(file2, []byte("b-original"), 0o644)

	s := &Snapshotter{Now: fixedTime}
	manifest, err := s.Create(snapshotDir, []string{file1, file2, file3})
	if err != nil {
		t.Fatal(err)
	}

	// Simulate modifications.
	os.WriteFile(file1, []byte("a-modified"), 0o644)
	os.WriteFile(file2, []byte("b-modified"), 0o644)
	os.WriteFile(file3, []byte("c-created"), 0o644)

	// Restore all.
	rs := NewRestoreService()
	if err := rs.Restore(manifest); err != nil {
		t.Fatal(err)
	}

	// Verify.
	d1, _ := os.ReadFile(file1)
	if string(d1) != "a-original" {
		t.Fatalf("file1 = %q", d1)
	}
	d2, _ := os.ReadFile(file2)
	if string(d2) != "b-original" {
		t.Fatalf("file2 = %q", d2)
	}
	if _, err := os.Stat(file3); !os.IsNotExist(err) {
		t.Fatal("file3 should have been removed")
	}
}
