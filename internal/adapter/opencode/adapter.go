// Package opencode implements the adapter for OpenCode.
package opencode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MiguelAguiarDEV/op-setup/internal/config"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Adapter manages OpenCode's configuration.
type Adapter struct {
	LookPath  func(string) (string, error)
	StatPath  func(string) (os.FileInfo, error)
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

// NewAdapter creates an OpenCode adapter with default dependencies.
func NewAdapter() *Adapter {
	return &Adapter{
		LookPath:  exec.LookPath,
		StatPath:  os.Stat,
		ReadFile:  os.ReadFile,
		WriteFile: config.WriteFileAtomic,
	}
}

func (a *Adapter) Name() string                   { return "OpenCode" }
func (a *Adapter) Agent() model.AgentID           { return model.AgentOpenCode }
func (a *Adapter) MCPStrategy() model.MCPStrategy { return model.StrategyMergeIntoJSON }
func (a *Adapter) MCPConfigKey() string           { return "mcp" }

func (a *Adapter) ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

func (a *Adapter) Detect(homeDir string) (model.DetectResult, error) {
	result := model.DetectResult{
		ConfigPath: a.ConfigPath(homeDir),
	}

	binPath, err := a.LookPath("opencode")
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

// PostInject adds "context-mode" to the "plugin" array in opencode.json
// if context-mode is among the selected components.
func (a *Adapter) PostInject(homeDir string, components []model.ComponentID) error {
	needsPlugin := false
	for _, c := range components {
		if c == model.ComponentContextMode {
			needsPlugin = true
			break
		}
	}
	if !needsPlugin {
		return nil
	}

	cfgPath := a.ConfigPath(homeDir)
	data, err := a.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Config doesn't exist yet — write a minimal config with the plugin entry.
			cfg := map[string]any{
				"plugin": []any{"context-mode"},
			}
			out, _ := json.MarshalIndent(cfg, "", "  ")
			out = append(out, '\n')
			return a.WriteFile(cfgPath, out, 0o644)
		}
		return err
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("config file is not valid: %s (%s)", cfgPath, err.Error())
	}

	// Get or create the plugin array.
	var plugins []any
	if raw, ok := cfg["plugin"]; ok {
		if arr, ok := raw.([]any); ok {
			plugins = arr
		}
	}

	// Check if context-mode already exists (idempotent).
	for _, p := range plugins {
		if s, ok := p.(string); ok && s == "context-mode" {
			return nil // Already present, no change needed.
		}
	}

	plugins = append(plugins, "context-mode")
	cfg["plugin"] = plugins

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return a.WriteFile(cfgPath, out, 0o644)
}
