package pipeline

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline/steps"
)

// Planner builds a StagePlan from user selections.
type Planner struct {
	Registry   *adapter.Registry
	HomeDir    string
	BackupRoot string // e.g., ~/.op-setup/backups/
}

// NewPlanner creates a Planner.
func NewPlanner(registry *adapter.Registry, homeDir string) *Planner {
	return &Planner{
		Registry:   registry,
		HomeDir:    homeDir,
		BackupRoot: filepath.Join(homeDir, ".op-setup", "backups"),
	}
}

// Plan creates a StagePlan for the given agent + component selections.
//
// Prepare steps: ValidateStep, BackupStep
// Apply steps: one InjectStep per selected agent
func (p *Planner) Plan(agents []model.AgentID, components []model.ComponentID) (StagePlan, error) {
	if len(agents) == 0 {
		return StagePlan{}, fmt.Errorf("no agents selected")
	}
	if len(components) == 0 {
		return StagePlan{}, fmt.Errorf("no components selected")
	}

	// Resolve adapters.
	var adapters []adapter.Adapter
	for _, agentID := range agents {
		a, ok := p.Registry.Get(agentID)
		if !ok {
			return StagePlan{}, fmt.Errorf("adapter not found for agent %q", agentID)
		}
		adapters = append(adapters, a)
	}

	// Resolve components.
	var comps []component.Component
	for _, compID := range components {
		c, ok := component.ByID(compID)
		if !ok {
			return StagePlan{}, fmt.Errorf("component not found: %q", compID)
		}
		comps = append(comps, c)
	}

	// Build dependency checks.
	var depChecks []steps.DependencyCheck
	seenBinaries := make(map[string]bool)

	for _, comp := range comps {
		if comp.Config.Type == model.MCPTypeLocal && len(comp.Config.Command) > 0 {
			binary := comp.Config.Command[0]
			if !seenBinaries[binary] {
				seenBinaries[binary] = true
				depChecks = append(depChecks, steps.DependencyCheck{
					Binary:   binary,
					Required: false, // Warn, don't fail — user may install later.
					Message:  fmt.Sprintf("required by %s", comp.Name),
				})
			}
		}
	}

	// Collect config paths for backup.
	var configPaths []string
	for _, a := range adapters {
		configPaths = append(configPaths, a.ConfigPath(p.HomeDir))
	}

	// Build Prepare steps.
	validateStep := steps.NewValidateStep(depChecks)

	backupDir := filepath.Join(p.BackupRoot, time.Now().Format("20060102-150405"))
	snapshotter := backup.NewSnapshotter()
	backupStep := steps.NewBackupStep(snapshotter, configPaths, backupDir)

	// Build Apply steps.
	resolver := component.NewResolver()
	var applySteps []Step
	for _, a := range adapters {
		injectStep := steps.NewInjectStep(a, comps, p.HomeDir, resolver)
		// The manifest will be set after backup runs.
		// We use a closure-like pattern: the orchestrator runs prepare first,
		// then we can access the manifest.
		applySteps = append(applySteps, &deferredManifestStep{
			inject:     injectStep,
			backupStep: backupStep,
		})
	}

	return StagePlan{
		Prepare: []Step{validateStep, backupStep},
		Apply:   applySteps,
	}, nil
}

// deferredManifestStep wraps an InjectStep and sets the manifest from
// the BackupStep before running.
type deferredManifestStep struct {
	inject     *steps.InjectStep
	backupStep *steps.BackupStep
}

func (s *deferredManifestStep) ID() string { return s.inject.ID() }

func (s *deferredManifestStep) Run() error {
	// Set the manifest from the backup step (which has already run in Prepare).
	if m := s.backupStep.Manifest(); m != nil {
		s.inject.SetManifest(m)
	}
	return s.inject.Run()
}

func (s *deferredManifestStep) Rollback() error {
	return s.inject.Rollback()
}
