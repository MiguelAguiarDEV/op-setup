package app

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

// headlessDeps holds injectable dependencies for the headless runner.
// Unexported — only used internally and in tests (same package).
type headlessDeps struct {
	homeDir           string
	adapterRegistry   *adapter.Registry
	installerRegistry *installer.Registry
	stdout            io.Writer
}

// RunHeadless runs the setup pipeline without the TUI.
// It auto-detects agents, selects components with satisfied env vars,
// and executes (or dry-runs) the pipeline.
func RunHeadless(cfg RunConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	registry, err := adapter.NewDefaultRegistry()
	if err != nil {
		return fmt.Errorf("create adapter registry: %w", err)
	}

	installerReg, err := installer.NewDefaultRegistry(homeDir)
	if err != nil {
		log.Printf("warning: installer registry unavailable: %v", err)
		installerReg = nil
	}

	return runHeadless(cfg, headlessDeps{
		homeDir:           homeDir,
		adapterRegistry:   registry,
		installerRegistry: installerReg,
		stdout:            os.Stdout,
	})
}

// runHeadless is the testable core of headless execution.
// All I/O goes through deps.stdout; all registries come from deps.
func runHeadless(cfg RunConfig, deps headlessDeps) error {
	profile := cfg.Profile

	// --- Detect agents ---
	var selectedAgents []model.AgentID
	if profile != model.ProfileDotfilesOnly {
		for _, a := range deps.adapterRegistry.All() {
			r, _ := a.Detect(deps.homeDir)
			if r.Installed {
				selectedAgents = append(selectedAgents, a.Agent())
				fmt.Fprintf(deps.stdout, "  detected: %s\n", a.Name())
			}
		}
		if len(selectedAgents) == 0 {
			return fmt.Errorf("no AI tools detected — install at least one (claude, opencode, codex, gemini)")
		}
	}

	// --- Select components ---
	var selectedComponents []model.ComponentID
	if profile != model.ProfileDotfilesOnly {
		for _, c := range component.All() {
			if component.EnvSatisfied(c) {
				selectedComponents = append(selectedComponents, c.ID)
			}
		}
		if len(selectedComponents) == 0 {
			return fmt.Errorf("no components have satisfied env vars — set required environment variables")
		}
	}

	// --- Build plan ---
	plan, err := pipeline.BuildPlan(deps.adapterRegistry, deps.installerRegistry, deps.homeDir, profile, selectedAgents, selectedComponents)
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}

	// --- Dry-run: print plan and exit ---
	if cfg.DryRun {
		fprintPlan(deps.stdout, plan, profile)
		return nil
	}

	// --- Execute ---
	fmt.Fprintf(deps.stdout, "\nExecuting %s...\n\n", profile.Name())

	orchestrator := pipeline.NewOrchestrator(func(e pipeline.ProgressEvent) {
		fprintProgressEvent(deps.stdout, e)
	})

	result := orchestrator.Execute(plan)

	fmt.Fprintln(deps.stdout)
	fprintResult(deps.stdout, result)

	if result.Err != nil {
		return result.Err
	}
	return nil
}

// fprintPlan writes the dry-run plan summary to w.
func fprintPlan(w io.Writer, plan pipeline.StagePlan, profile model.SetupProfile) {
	fmt.Fprintf(w, "\n[DRY RUN] Plan for profile: %s\n", profile.Name())

	stages := []struct {
		name  string
		steps []pipeline.Step
	}{
		{"Prepare", plan.Prepare},
		{"Install", plan.Install},
		{"Deploy", plan.Deploy},
		{"Apply", plan.Apply},
	}

	total := 0
	for _, s := range stages {
		if len(s.steps) == 0 {
			continue
		}
		fmt.Fprintf(w, "  %s:\n", s.name)
		for _, step := range s.steps {
			fmt.Fprintf(w, "    - %s\n", step.ID())
			total++
		}
	}
	fmt.Fprintf(w, "Total: %d steps\n", total)
}

// fprintProgressEvent writes a single progress event to w.
func fprintProgressEvent(w io.Writer, e pipeline.ProgressEvent) {
	var icon string
	switch e.Status {
	case pipeline.StatusRunning:
		icon = "..."
	case pipeline.StatusSucceeded:
		icon = " ok"
	case pipeline.StatusFailed:
		icon = "ERR"
	case pipeline.StatusRolledBack:
		icon = "<--"
	default:
		icon = "   "
	}

	line := fmt.Sprintf("  [%s] [%s] %s", e.Stage, icon, e.StepID)
	if e.Err != nil {
		line += fmt.Sprintf(" — %v", e.Err)
	}
	fmt.Fprintln(w, line)
}

// fprintResult writes the execution result summary to w.
func fprintResult(w io.Writer, result pipeline.ExecutionResult) {
	if result.Err == nil {
		fmt.Fprintln(w, "Setup complete!")
	} else {
		fmt.Fprintf(w, "Setup failed: %v\n", result.Err)
	}

	if result.Rollback != nil {
		fmt.Fprintln(w, "Rollback was executed.")
		for _, sr := range result.Rollback.Steps {
			status := "ok"
			if sr.Status == pipeline.StatusFailed {
				status = "FAILED"
			}
			fmt.Fprintf(w, "  [%s] %s\n", status, sr.StepID)
		}
	}
}
