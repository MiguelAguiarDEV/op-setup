// Package app wires together all components and runs the application.
package app

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Run initializes and runs the application.
// If cfg.NonInteractive is true, runs headless; otherwise launches the TUI.
func Run(cfg RunConfig) error {
	if cfg.NonInteractive {
		return RunHeadless(cfg)
	}
	return runTUI(cfg)
}

// runTUI launches the interactive TUI.
func runTUI(cfg RunConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	registry, err := adapter.NewDefaultRegistry()
	if err != nil {
		return fmt.Errorf("create adapter registry: %w", err)
	}

	// Create installer registry. Non-fatal if it fails — ProfileFull
	// will simply have no install steps.
	installerReg, err := installer.NewDefaultRegistry(homeDir)
	if err != nil {
		log.Printf("warning: installer registry unavailable: %v", err)
		installerReg = nil
	}

	// ProgramRef allows the model to send messages from goroutines.
	// The reference is set after tea.NewProgram() creates the program.
	ref := &tui.ProgramRef{}

	// Build model options from CLI config.
	var opts []tui.ModelOption
	if cfg.DryRun {
		opts = append(opts, tui.WithDryRun(true))
	}
	if cfg.Profile != "" {
		opts = append(opts, tui.WithProfile(cfg.Profile))
	}

	m := tui.NewModel(registry, installerReg, ref, Version, homeDir, opts...)
	p := tea.NewProgram(m, tea.WithAltScreen())
	ref.P = p

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}

	return nil
}
