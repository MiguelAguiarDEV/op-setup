package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/config"
	"github.com/MiguelAguiarDEV/op-setup/internal/dotfiles"
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline/steps"
)

// Planner builds a StagePlan from user selections and profile.
type Planner struct {
	Registry          *adapter.Registry
	InstallerRegistry *installer.Registry
	HomeDir           string
	BackupRoot        string
}

// NewPlanner creates a Planner with default backup root.
func NewPlanner(registry *adapter.Registry, homeDir string) *Planner {
	return &Planner{
		Registry:   registry,
		HomeDir:    homeDir,
		BackupRoot: filepath.Join(homeDir, ".op-setup", "backups"),
	}
}

// Plan creates a StagePlan for the given profile and selections.
//
// Profile determines which stages are populated:
//   - ProfileFull:         Prepare + Install + Deploy + Apply
//   - ProfileMCPOnly:      Prepare + Apply
//   - ProfileDotfilesOnly: Prepare + Deploy
func (p *Planner) Plan(
	profile model.SetupProfile,
	agents []model.AgentID,
	components []model.ComponentID,
) (StagePlan, error) {
	plan := StagePlan{}
	timestamp := time.Now().Format("20060102-150405")

	switch profile {
	case model.ProfileFull:
		return p.planFull(agents, components, timestamp)
	case model.ProfileMCPOnly:
		return p.planMCPOnly(agents, components, timestamp)
	case model.ProfileDotfilesOnly:
		return p.planDotfilesOnly(timestamp)
	default:
		return plan, fmt.Errorf("unsupported profile: %q", profile)
	}
}

// PlanMCP creates a StagePlan for MCP-only mode (original v1 behavior).
// This is a convenience method that preserves backward compatibility.
func (p *Planner) PlanMCP(agents []model.AgentID, components []model.ComponentID) (StagePlan, error) {
	return p.Plan(model.ProfileMCPOnly, agents, components)
}

func (p *Planner) planFull(
	agents []model.AgentID,
	components []model.ComponentID,
	timestamp string,
) (StagePlan, error) {
	plan := StagePlan{}

	if len(agents) == 0 {
		return plan, fmt.Errorf("no agents selected")
	}
	if len(components) == 0 {
		return plan, fmt.Errorf("no components selected")
	}

	// --- Prepare ---
	prepareSteps, backupStep, err := p.buildPrepareSteps(agents, components, timestamp)
	if err != nil {
		return plan, err
	}
	plan.Prepare = prepareSteps

	// --- Install ---
	if p.InstallerRegistry != nil {
		plan.Install = p.buildInstallSteps()
	}

	// --- Deploy ---
	plan.Deploy = p.buildDeploySteps(timestamp)

	// --- Apply ---
	applySteps, err := p.buildApplySteps(agents, components, backupStep)
	if err != nil {
		return plan, err
	}
	plan.Apply = applySteps

	return plan, nil
}

func (p *Planner) planMCPOnly(
	agents []model.AgentID,
	components []model.ComponentID,
	timestamp string,
) (StagePlan, error) {
	plan := StagePlan{}

	if len(agents) == 0 {
		return plan, fmt.Errorf("no agents selected")
	}
	if len(components) == 0 {
		return plan, fmt.Errorf("no components selected")
	}

	prepareSteps, backupStep, err := p.buildPrepareSteps(agents, components, timestamp)
	if err != nil {
		return plan, err
	}
	plan.Prepare = prepareSteps

	applySteps, err := p.buildApplySteps(agents, components, backupStep)
	if err != nil {
		return plan, err
	}
	plan.Apply = applySteps

	return plan, nil
}

func (p *Planner) planDotfilesOnly(timestamp string) (StagePlan, error) {
	plan := StagePlan{}

	// Minimal prepare: no dependency checks needed for dotfiles.
	plan.Prepare = []Step{}

	// Deploy dotfiles.
	plan.Deploy = p.buildDeploySteps(timestamp)

	return plan, nil
}

// buildPrepareSteps creates ValidateStep + BackupStep.
func (p *Planner) buildPrepareSteps(
	agents []model.AgentID,
	components []model.ComponentID,
	timestamp string,
) ([]Step, *steps.BackupStep, error) {
	adapters, err := p.resolveAdapters(agents)
	if err != nil {
		return nil, nil, err
	}

	comps, err := p.resolveComponents(components)
	if err != nil {
		return nil, nil, err
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
					Required: false,
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

	validateStep := steps.NewValidateStep(depChecks)
	backupDir := filepath.Join(p.BackupRoot, timestamp)
	snapshotter := backup.NewSnapshotter()
	backupStep := steps.NewBackupStep(snapshotter, configPaths, backupDir)

	return []Step{validateStep, backupStep}, backupStep, nil
}

// buildInstallSteps creates an InstallStep for each registered installer.
func (p *Planner) buildInstallSteps() []Step {
	if p.InstallerRegistry == nil {
		return nil
	}
	all := p.InstallerRegistry.All()
	installSteps := make([]Step, len(all))
	for i, inst := range all {
		installSteps[i] = &installer.InstallStep{
			Installer: inst,
			Ctx:       context.Background(),
		}
	}
	return installSteps
}

// buildDeploySteps creates a DeployStep for dotfiles.
func (p *Planner) buildDeploySteps(timestamp string) []Step {
	configDir := filepath.Join(p.HomeDir, ".config")
	snapshotDir := filepath.Join(p.BackupRoot, timestamp+"-dotfiles")

	deployer := &dotfiles.Deployer{
		FS:              dotfiles.EmbeddedFS,
		ConfigDir:       configDir,
		WriteFileAtomic: config.WriteFileAtomic,
		ReadFile:        os.ReadFile,
		MkdirAll:        os.MkdirAll,
	}

	return []Step{&dotfiles.DeployStep{
		Deployer:    deployer,
		SnapshotDir: snapshotDir,
	}}
}

// buildApplySteps creates InjectSteps for MCP config injection.
func (p *Planner) buildApplySteps(
	agents []model.AgentID,
	components []model.ComponentID,
	backupStep *steps.BackupStep,
) ([]Step, error) {
	adapters, err := p.resolveAdapters(agents)
	if err != nil {
		return nil, err
	}

	comps, err := p.resolveComponents(components)
	if err != nil {
		return nil, err
	}

	resolver := component.NewResolver()
	var applySteps []Step
	for _, a := range adapters {
		injectStep := steps.NewInjectStep(a, comps, p.HomeDir, resolver)
		applySteps = append(applySteps, &deferredManifestStep{
			inject:     injectStep,
			backupStep: backupStep,
		})
	}

	return applySteps, nil
}

func (p *Planner) resolveAdapters(agents []model.AgentID) ([]adapter.Adapter, error) {
	var adapters []adapter.Adapter
	for _, agentID := range agents {
		a, ok := p.Registry.Get(agentID)
		if !ok {
			return nil, fmt.Errorf("adapter not found for agent %q", agentID)
		}
		adapters = append(adapters, a)
	}
	return adapters, nil
}

func (p *Planner) resolveComponents(components []model.ComponentID) ([]component.Component, error) {
	var comps []component.Component
	for _, compID := range components {
		c, ok := component.ByID(compID)
		if !ok {
			return nil, fmt.Errorf("component not found: %q", compID)
		}
		comps = append(comps, c)
	}
	return comps, nil
}

// Compile-time interface check: deferredManifestStep must implement RollbackStep.
var _ RollbackStep = (*deferredManifestStep)(nil)

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
