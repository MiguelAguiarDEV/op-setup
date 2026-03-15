// Package backup provides config file backup and restore functionality.
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Manifest describes a backup snapshot.
type Manifest struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	RootDir   string          `json:"root_dir"`
	Entries   []ManifestEntry `json:"entries"`
}

// ManifestEntry describes one backed-up file.
type ManifestEntry struct {
	// OriginalPath is the absolute path to the original file.
	OriginalPath string `json:"original_path"`

	// SnapshotPath is the path to the backup copy within the snapshot dir.
	SnapshotPath string `json:"snapshot_path"`

	// Existed is true if the file existed before backup.
	// If false, rollback should remove the file.
	Existed bool `json:"existed"`

	// Mode is the original file permissions (only set if Existed=true).
	Mode uint32 `json:"mode,omitempty"`
}

// WriteManifest writes a manifest to the given path as JSON.
func WriteManifest(path string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	data = append(data, '\n')

	dir := path[:len(path)-len("/manifest.json")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// ReadManifest reads a manifest from the given path.
func ReadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}

	return m, nil
}
