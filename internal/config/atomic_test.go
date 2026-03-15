package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomic_WritesCorrectly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key": "value"}`)

	if err := WriteFileAtomic(path, data, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("content = %q, want %q", got, data)
	}
}

func TestWriteFileAtomic_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "test.json")
	data := []byte(`{}`)

	if err := WriteFileAtomic(path, data, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestWriteFileAtomic_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{}`)

	if err := WriteFileAtomic(path, data, 0o600); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// Check that the permission bits match (ignoring umask effects on some systems).
	got := info.Mode().Perm()
	if got != 0o600 {
		t.Fatalf("permissions = %o, want %o", got, 0o600)
	}
}

func TestWriteFileAtomic_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// Write initial content.
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Overwrite atomically.
	newData := []byte("new content")
	if err := WriteFileAtomic(path, newData, 0o644); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "new content" {
		t.Fatalf("content = %q, want %q", got, "new content")
	}
}

func TestWriteFileAtomic_NoTempFileLeftOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := WriteFileAtomic(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "test.json" {
			t.Fatalf("unexpected file left behind: %s", e.Name())
		}
	}
}
