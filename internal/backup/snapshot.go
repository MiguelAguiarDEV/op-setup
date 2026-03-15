package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Snapshotter creates backup snapshots of config files.
type Snapshotter struct {
	// Now returns the current time. Injectable for testing.
	Now func() time.Time
}

// NewSnapshotter creates a Snapshotter with default dependencies.
func NewSnapshotter() *Snapshotter {
	return &Snapshotter{
		Now: time.Now,
	}
}

// Create creates a backup snapshot of the given file paths.
// Files are copied into snapshotDir/files/... and a manifest.json is written.
// Returns the manifest describing the snapshot.
//
// If a source file doesn't exist, its ManifestEntry will have Existed=false.
func (s *Snapshotter) Create(snapshotDir string, paths []string) (Manifest, error) {
	now := s.Now()
	id := now.Format("20060102-150405")

	filesDir := filepath.Join(snapshotDir, "files")
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		return Manifest{}, fmt.Errorf("create snapshot dir: %w", err)
	}

	manifest := Manifest{
		ID:        id,
		CreatedAt: now,
		RootDir:   snapshotDir,
		Entries:   make([]ManifestEntry, 0, len(paths)),
	}

	for i, srcPath := range paths {
		entry := ManifestEntry{
			OriginalPath: srcPath,
		}

		info, err := os.Stat(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				entry.Existed = false
				entry.SnapshotPath = "" // No backup file for non-existent sources.
				manifest.Entries = append(manifest.Entries, entry)
				continue
			}
			return Manifest{}, fmt.Errorf("stat %s: %w", srcPath, err)
		}

		entry.Existed = true
		entry.Mode = uint32(info.Mode().Perm())

		// Use index-based naming to avoid path conflicts.
		dstName := fmt.Sprintf("%d_%s", i, filepath.Base(srcPath))
		dstPath := filepath.Join(filesDir, dstName)
		entry.SnapshotPath = dstPath

		if err := copyFile(srcPath, dstPath); err != nil {
			return Manifest{}, fmt.Errorf("copy %s: %w", srcPath, err)
		}

		manifest.Entries = append(manifest.Entries, entry)
	}

	// Write manifest.
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	if err := WriteManifest(manifestPath, manifest); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

// copyFile copies src to dst, preserving permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	info, err := in.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode().Perm())
}
