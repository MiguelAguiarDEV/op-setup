package backup

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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
			if err := os.Remove(entry.OriginalPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("remove %s: %w", entry.OriginalPath, err)
			}
		}
	}
	return nil
}

// restoreFile copies src to dst with the given permissions.
func restoreFile(src, dst string, perm os.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, io.LimitReader(in, maxConfigSize)); err != nil {
		return err
	}

	return nil
}
