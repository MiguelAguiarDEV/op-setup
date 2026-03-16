package dotfiles

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FileAction describes what will happen to a file during deployment.
type FileAction string

const (
	// ActionCreate means the file does not exist and will be created.
	ActionCreate FileAction = "create"

	// ActionOverwrite means the file exists with different content and will be overwritten.
	ActionOverwrite FileAction = "overwrite"

	// ActionSkip means the file exists with identical content and will be skipped.
	ActionSkip FileAction = "skip"
)

// PlanEntry describes the planned action for a single file.
type PlanEntry struct {
	Mapping FileMapping
	Action  FileAction
}

// DeployResult describes the outcome of deploying a single file.
type DeployResult struct {
	Mapping FileMapping
	Action  FileAction
	Err     error
}

// Deployer plans and executes dotfile deployments.
type Deployer struct {
	// FS is the embedded filesystem containing dotfiles.
	FS fs.FS

	// ConfigDir is the base config directory (e.g. $XDG_CONFIG_HOME or ~/.config).
	ConfigDir string

	// WriteFileAtomic writes a file atomically. Defaults to config.WriteFileAtomic.
	WriteFileAtomic func(path string, data []byte, perm os.FileMode) error

	// ReadFile reads a file from disk. Defaults to os.ReadFile.
	ReadFile func(string) ([]byte, error)

	// MkdirAll creates directories. Defaults to os.MkdirAll.
	MkdirAll func(string, os.FileMode) error

	// backedUpPaths tracks paths that were backed up for rollback.
	backedUpPaths []string
}

// Plan scans all embedded files and determines what action each requires.
func (d *Deployer) Plan() ([]PlanEntry, error) {
	mappings, err := BuildManifest(d.FS, d.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("build manifest: %w", err)
	}

	readFile := d.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	entries := make([]PlanEntry, 0, len(mappings))

	for _, m := range mappings {
		embeddedData, err := fs.ReadFile(d.FS, m.EmbedPath)
		if err != nil {
			return nil, &ReadEmbedError{Path: m.EmbedPath, Reason: err.Error()}
		}

		existing, readErr := readFile(m.TargetPath)
		if readErr != nil {
			if !errors.Is(readErr, fs.ErrNotExist) {
				return nil, fmt.Errorf("read target %s: %w", m.TargetPath, readErr)
			}
			// File doesn't exist → create.
			entries = append(entries, PlanEntry{
				Mapping: m,
				Action:  ActionCreate,
			})
			continue
		}

		if bytes.Equal(existing, embeddedData) {
			entries = append(entries, PlanEntry{
				Mapping: m,
				Action:  ActionSkip,
			})
		} else {
			entries = append(entries, PlanEntry{
				Mapping: m,
				Action:  ActionOverwrite,
			})
		}
	}

	return entries, nil
}

// Deploy executes the deployment plan. It creates directories, backs up existing
// files (via the provided backupFunc), and writes embedded files atomically.
//
// backupFunc is called with the list of paths that will be overwritten, before
// any writes happen. It should create backups and return an error if backup fails.
// Pass nil to skip backups.
func (d *Deployer) Deploy(plan []PlanEntry, backupFunc func(paths []string) error) ([]DeployResult, error) {
	writeAtomic := d.WriteFileAtomic
	if writeAtomic == nil {
		return nil, fmt.Errorf("WriteFileAtomic is required")
	}

	mkdirAll := d.MkdirAll
	if mkdirAll == nil {
		mkdirAll = os.MkdirAll
	}

	// Collect paths that need backup (overwrite only).
	var overwritePaths []string
	for _, entry := range plan {
		if entry.Action == ActionOverwrite {
			overwritePaths = append(overwritePaths, entry.Mapping.TargetPath)
		}
	}

	// Backup before any writes.
	if backupFunc != nil && len(overwritePaths) > 0 {
		if err := backupFunc(overwritePaths); err != nil {
			return nil, fmt.Errorf("backup before deploy: %w", err)
		}
		d.backedUpPaths = overwritePaths
	}

	results := make([]DeployResult, 0, len(plan))

	for _, entry := range plan {
		if entry.Action == ActionSkip {
			results = append(results, DeployResult{
				Mapping: entry.Mapping,
				Action:  ActionSkip,
			})
			continue
		}

		// Read embedded content.
		data, err := fs.ReadFile(d.FS, entry.Mapping.EmbedPath)
		if err != nil {
			results = append(results, DeployResult{
				Mapping: entry.Mapping,
				Action:  entry.Action,
				Err:     &ReadEmbedError{Path: entry.Mapping.EmbedPath, Reason: err.Error()},
			})
			continue
		}

		// Ensure parent directory exists.
		dir := filepath.Dir(entry.Mapping.TargetPath)
		if err := mkdirAll(dir, 0o700); err != nil {
			results = append(results, DeployResult{
				Mapping: entry.Mapping,
				Action:  entry.Action,
				Err:     &DeployFailedError{Path: entry.Mapping.TargetPath, Reason: err.Error()},
			})
			continue
		}

		// Write atomically.
		if err := writeAtomic(entry.Mapping.TargetPath, data, 0o600); err != nil {
			results = append(results, DeployResult{
				Mapping: entry.Mapping,
				Action:  entry.Action,
				Err:     &DeployFailedError{Path: entry.Mapping.TargetPath, Reason: err.Error()},
			})
			continue
		}

		results = append(results, DeployResult{
			Mapping: entry.Mapping,
			Action:  entry.Action,
		})
	}

	return results, nil
}

// BackedUpPaths returns the paths that were backed up during the last Deploy call.
func (d *Deployer) BackedUpPaths() []string {
	return d.backedUpPaths
}
