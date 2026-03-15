package dotfiles

import (
	"io/fs"
	"path/filepath"
)

// FileMapping describes a single file to deploy.
type FileMapping struct {
	// EmbedPath is the path within the embedded FS (e.g. "embed/opencode/AGENTS.md").
	EmbedPath string

	// TargetPath is the absolute destination path on disk.
	TargetPath string
}

// BuildManifest walks the embedded FS and builds a list of file mappings.
// Each embedded file under "embed/opencode/" maps to configDir/opencode/...
// and each file under "embed/nvim/" maps to configDir/nvim/...
//
// configDir is typically ~/.config.
func BuildManifest(embeddedFS fs.FS, configDir string) ([]FileMapping, error) {
	var mappings []FileMapping

	err := fs.WalkDir(embeddedFS, "embed", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Strip the "embed/" prefix to get the relative path.
		// e.g. "embed/opencode/AGENTS.md" → "opencode/AGENTS.md"
		relPath := path[len("embed/"):]

		targetPath := filepath.Join(configDir, relPath)

		mappings = append(mappings, FileMapping{
			EmbedPath:  path,
			TargetPath: targetPath,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return mappings, nil
}
