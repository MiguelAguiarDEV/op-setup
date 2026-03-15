// Package claude implements the adapter for Claude Code.
package claude

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Adapter manages Claude Code's configuration.
type Adapter struct {
	// LookPath resolves a binary name to its path. Defaults to exec.LookPath.
	LookPath func(string) (string, error)

	// StatPath checks if a path exists. Defaults to os.Stat.
	StatPath func(string) (os.FileInfo, error)
}

// NewAdapter creates a Claude Code adapter with default dependencies.
func NewAdapter() *Adapter {
	return &Adapter{
		LookPath: exec.LookPath,
		StatPath: os.Stat,
	}
}

func (a *Adapter) Name() string                   { return "Claude Code" }
func (a *Adapter) Agent() model.AgentID           { return model.AgentClaudeCode }
func (a *Adapter) MCPStrategy() model.MCPStrategy { return model.StrategyMergeIntoJSON }
func (a *Adapter) MCPConfigKey() string           { return "mcpServers" }

func (a *Adapter) ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".claude", "settings.json")
}

func (a *Adapter) Detect(homeDir string) (model.DetectResult, error) {
	result := model.DetectResult{
		ConfigPath: a.ConfigPath(homeDir),
	}

	binPath, err := a.LookPath("claude")
	if err == nil {
		result.Installed = true
		result.BinaryPath = binPath
	}

	_, err = a.StatPath(result.ConfigPath)
	if err == nil {
		result.ConfigFound = true
	} else if !os.IsNotExist(err) {
		return result, err
	}

	return result, nil
}

// PostInject is a no-op for Claude Code.
func (a *Adapter) PostInject(_ string, _ []model.ComponentID) error {
	return nil
}
