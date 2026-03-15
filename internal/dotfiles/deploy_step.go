package dotfiles

import (
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
)

// DeployStep wraps a Deployer to implement pipeline.Step and pipeline.RollbackStep.
// Compile-time interface check is in deploy_step_check_test.go to avoid import cycle.
type DeployStep struct {
	Deployer    *Deployer
	SnapshotDir string
	manifest    *backup.Manifest
	deployed    bool
}

// ID returns the step identifier.
func (s *DeployStep) ID() string {
	return "deploy-dotfiles"
}

// Run plans and deploys all embedded dotfiles.
func (s *DeployStep) Run() error {
	plan, err := s.Deployer.Plan()
	if err != nil {
		return fmt.Errorf("plan dotfiles: %w", err)
	}

	// Count actionable items.
	actionable := 0
	for _, entry := range plan {
		if entry.Action != ActionSkip {
			actionable++
		}
	}
	if actionable == 0 {
		return nil // Nothing to deploy.
	}

	// Deploy with backup.
	backupFunc := func(paths []string) error {
		snapshotter := backup.NewSnapshotter()
		m, err := snapshotter.Create(s.SnapshotDir, paths)
		if err != nil {
			return err
		}
		s.manifest = &m
		return nil
	}

	results, err := s.Deployer.Deploy(plan, backupFunc)
	if err != nil {
		return err
	}

	// Check for individual file errors.
	for _, r := range results {
		if r.Err != nil {
			return fmt.Errorf("deploy %s: %w", r.Mapping.TargetPath, r.Err)
		}
	}

	s.deployed = true
	return nil
}

// Rollback restores backed-up files from the snapshot.
func (s *DeployStep) Rollback() error {
	if !s.deployed || s.manifest == nil {
		return nil
	}
	rs := backup.NewRestoreService()
	return rs.Restore(*s.manifest)
}

// Deployed returns true if files were actually deployed.
func (s *DeployStep) Deployed() bool {
	return s.deployed
}
