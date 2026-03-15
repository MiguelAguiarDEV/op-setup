package installer_test

import (
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

// Compile-time interface check: InstallStep must implement pipeline.RollbackStep.
// This is in a _test.go file to avoid an import cycle (pipeline → installer → pipeline).
var _ pipeline.RollbackStep = (*installer.InstallStep)(nil)
