package dotfiles_test

import (
	"github.com/MiguelAguiarDEV/op-setup/internal/dotfiles"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

// Compile-time interface check: DeployStep must implement pipeline.RollbackStep.
// This is in a _test.go file to avoid an import cycle (pipeline → dotfiles → pipeline).
var _ pipeline.RollbackStep = (*dotfiles.DeployStep)(nil)
