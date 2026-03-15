// Package codex implements the adapter for Codex.
package codex

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Adapter manages Codex's configuration.
type Adapter struct {
	LookPath func(string) (string, error)
	StatPath func(string) (os.FileInfo, error)
}

// NewAdapter creates a Codex adapter with default dependencies.
func NewAdapter() *Adapter {
	return &Adapter{
		LookPath: exec.LookPath,
		StatPath: os.Stat,
	}
}

func (a *Adapter) Name() string                   { return "Codex" }
func (a *Adapter) Agent() model.AgentID           { return model.AgentCodex }
func (a *Adapter) MCPStrategy() model.MCPStrategy { return model.StrategyMergeIntoTOML }
func (a *Adapter) MCPConfigKey() string           { return "mcp_servers" }

func (a *Adapter) ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".codex", "config.toml")
}

func (a *Adapter) Detect(homeDir string) (model.DetectResult, error) {
	result := model.DetectResult{
		ConfigPath: a.ConfigPath(homeDir),
	}

	binPath, err := a.LookPath("codex")
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

// PostInject is a no-op for Codex.
func (a *Adapter) PostInject(_ string, _ []model.ComponentID) error {
	return nil
}
