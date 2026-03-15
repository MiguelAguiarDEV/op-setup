package app

import (
	"fmt"
	"log"
	"os"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

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

	profile := cfg.Profile

	// --- Detect agents ---
	var selectedAgents []model.AgentID
	if profile != model.ProfileDotfilesOnly {
		for _, a := range registry.All() {
			r, _ := a.Detect(homeDir)
			if r.Installed {
				selectedAgents = append(selectedAgents, a.Agent())
				fmt.Printf("  detected: %s\n", a.Name())
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
	planner := pipeline.NewPlanner(registry, homeDir)
	if profile == model.ProfileFull && installerReg != nil {
		planner.InstallerRegistry = installerReg
	}

	var plan pipeline.StagePlan
	switch profile {
	case model.ProfileDotfilesOnly:
		plan, err = planner.Plan(profile, nil, nil)
	default:
		plan, err = planner.Plan(profile, selectedAgents, selectedComponents)
	}
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}

	// --- Dry-run: print plan and exit ---
	if cfg.DryRun {
		printPlan(plan, profile)
		return nil
	}

	// --- Execute ---
	fmt.Printf("\nExecuting %s...\n\n", profile.Name())

	orchestrator := pipeline.NewOrchestrator(func(e pipeline.ProgressEvent) {
		printProgressEvent(e)
	})

	result := orchestrator.Execute(plan)

	fmt.Println()
	printResult(result)

	if result.Err != nil {
		return result.Err
	}
	return nil
}

// printPlan prints the dry-run plan summary.
func printPlan(plan pipeline.StagePlan, profile model.SetupProfile) {
	fmt.Printf("\n[DRY RUN] Plan for profile: %s\n", profile.Name())

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
		fmt.Printf("  %s:\n", s.name)
		for _, step := range s.steps {
			fmt.Printf("    - %s\n", step.ID())
			total++
		}
	}
	fmt.Printf("Total: %d steps\n", total)
}

// printProgressEvent prints a single progress event to stdout.
func printProgressEvent(e pipeline.ProgressEvent) {
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
	fmt.Println(line)
}

// printResult prints the execution result summary.
func printResult(result pipeline.ExecutionResult) {
	if result.Err == nil {
		fmt.Println("Setup complete!")
	} else {
		fmt.Printf("Setup failed: %v\n", result.Err)
	}

	if result.Rollback != nil {
		fmt.Println("Rollback was executed.")
		for _, sr := range result.Rollback.Steps {
			status := "ok"
			if sr.Status == pipeline.StatusFailed {
				status = "FAILED"
			}
			fmt.Printf("  [%s] %s\n", status, sr.StepID)
		}
	}
}
