// Package app wires together all components and runs the TUI.
package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Run initializes and runs the TUI application.
func Run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	registry, err := adapter.NewDefaultRegistry()
	if err != nil {
		return fmt.Errorf("create adapter registry: %w", err)
	}

	m := tui.NewModel(registry, Version, homeDir)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}

	return nil
}
