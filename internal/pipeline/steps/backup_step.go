// Package steps provides concrete pipeline step implementations.
package steps

import (
	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
)

// BackupStep creates a backup of all config files that will be modified.
type BackupStep struct {
	snapshotter *backup.Snapshotter
	paths       []string
	backupDir   string
	manifest    *backup.Manifest
}

// NewBackupStep creates a BackupStep.
func NewBackupStep(snapshotter *backup.Snapshotter, paths []string, backupDir string) *BackupStep {
	return &BackupStep{
		snapshotter: snapshotter,
		paths:       paths,
		backupDir:   backupDir,
	}
}

func (s *BackupStep) ID() string { return "backup-configs" }

// Run creates a snapshot of all config files.
func (s *BackupStep) Run() error {
	m, err := s.snapshotter.Create(s.backupDir, s.paths)
	if err != nil {
		return err
	}
	s.manifest = &m
	return nil
}

// Manifest returns the backup manifest (populated after Run).
func (s *BackupStep) Manifest() *backup.Manifest {
	return s.manifest
}
