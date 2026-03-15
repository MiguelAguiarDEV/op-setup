// Package gemini implements the adapter for Gemini CLI (Antigravity).
package gemini

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Adapter manages Gemini CLI's configuration.
type Adapter struct {
	LookPath func(string) (string, error)
	StatPath func(string) (os.FileInfo, error)
}

// NewAdapter creates a Gemini CLI adapter with default dependencies.
func NewAdapter() *Adapter {
	return &Adapter{
		LookPath: exec.LookPath,
		StatPath: os.Stat,
	}
}

func (a *Adapter) Name() string                   { return "Gemini CLI" }
func (a *Adapter) Agent() model.AgentID           { return model.AgentGeminiCLI }
func (a *Adapter) MCPStrategy() model.MCPStrategy { return model.StrategyMergeIntoJSON }
func (a *Adapter) MCPConfigKey() string           { return "mcpServers" }

func (a *Adapter) ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "settings.json")
}

func (a *Adapter) Detect(homeDir string) (model.DetectResult, error) {
	result := model.DetectResult{
		ConfigPath: a.ConfigPath(homeDir),
	}

	binPath, err := a.LookPath("gemini")
	if err == nil {
		result.Installed = true
		result.BinaryPath = binPath
	}

	_, err = a.StatPath(result.ConfigPath)
	if err == nil {
		result.ConfigFound = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return result, err
	}

	return result, nil
}

// PostInject is a no-op for Gemini CLI.
func (a *Adapter) PostInject(_ string, _ []model.ComponentID) error {
	return nil
}
