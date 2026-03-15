package backup

import (
	"fmt"
	"io"
	"os"
)

// RestoreService restores files from a backup manifest.
type RestoreService struct{}

// NewRestoreService creates a new RestoreService.
func NewRestoreService() *RestoreService {
	return &RestoreService{}
}

// Restore restores all files described in the manifest.
//
// For entries where Existed=true, the backup copy is restored to the original path.
// For entries where Existed=false, the original path is removed (if it exists now).
func (s *RestoreService) Restore(manifest Manifest) error {
	for _, entry := range manifest.Entries {
		if entry.Existed {
			if entry.SnapshotPath == "" {
				return fmt.Errorf("restore %s: snapshot path is empty but file existed", entry.OriginalPath)
			}
			if err := restoreFile(entry.SnapshotPath, entry.OriginalPath, os.FileMode(entry.Mode)); err != nil {
				return fmt.Errorf("restore %s: %w", entry.OriginalPath, err)
			}
		} else {
			// File didn't exist before — remove it if it was created.
			if err := os.Remove(entry.OriginalPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", entry.OriginalPath, err)
			}
		}
	}
	return nil
}

// restoreFile copies src to dst with the given permissions.
func restoreFile(src, dst string, perm os.FileMode) error {
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

	return os.Chmod(dst, perm)
}
