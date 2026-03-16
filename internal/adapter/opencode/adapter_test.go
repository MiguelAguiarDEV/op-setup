package opencode

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestAdapter_Identity(t *testing.T) {
	a := NewAdapter()
	if a.Name() != "OpenCode" {
		t.Fatalf("Name() = %q, want %q", a.Name(), "OpenCode")
	}
	if a.Agent() != model.AgentOpenCode {
		t.Fatalf("Agent() = %q, want %q", a.Agent(), model.AgentOpenCode)
	}
	if a.MCPStrategy() != model.StrategyMergeIntoJSON {
		t.Fatalf("MCPStrategy() = %d, want %d", a.MCPStrategy(), model.StrategyMergeIntoJSON)
	}
	if a.MCPConfigKey() != "mcp" {
		t.Fatalf("MCPConfigKey() = %q, want %q", a.MCPConfigKey(), "mcp")
	}
}

func TestAdapter_ConfigPath_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	a := NewAdapter()
	got := a.ConfigPath("/home/test")
	want := filepath.Join("/home/test", ".config", "opencode", "opencode.json")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestAdapter_ConfigPath_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	a := NewAdapter()
	got := a.ConfigPath("/home/test")
	want := filepath.Join("/custom/config", "opencode", "opencode.json")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestAdapter_Detect(t *testing.T) {
	tests := []struct {
		name        string
		lookPath    func(string) (string, error)
		statPath    func(string) (os.FileInfo, error)
		wantInstall bool
		wantConfig  bool
		wantErr     bool
	}{
		{
			name:        "binary found and config exists",
			lookPath:    func(string) (string, error) { return "/usr/bin/opencode", nil },
			statPath:    func(string) (os.FileInfo, error) { return nil, nil },
			wantInstall: true,
			wantConfig:  true,
		},
		{
			name:        "binary not found and config missing",
			lookPath:    func(string) (string, error) { return "", errors.New("not found") },
			statPath:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantInstall: false,
			wantConfig:  false,
		},
		{
			name:     "stat error bubbles up",
			lookPath: func(string) (string, error) { return "/usr/bin/opencode", nil },
			statPath: func(string) (os.FileInfo, error) {
				return nil, errors.New("permission denied")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				LookPath: tt.lookPath,
				StatPath: tt.statPath,
			}
			result, err := a.Detect("/home/test")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Installed != tt.wantInstall {
				t.Fatalf("Installed = %v, want %v", result.Installed, tt.wantInstall)
			}
			if result.ConfigFound != tt.wantConfig {
				t.Fatalf("ConfigFound = %v, want %v", result.ConfigFound, tt.wantConfig)
			}
		})
	}
}

func TestAdapter_Detect_LookPathCalledWithOpencode(t *testing.T) {
	var calledWith string
	a := &Adapter{
		LookPath: func(name string) (string, error) {
			calledWith = name
			return "", errors.New("not found")
		},
		StatPath: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
	}
	_, _ = a.Detect("/home/test")
	if calledWith != "opencode" {
		t.Fatalf("LookPath called with %q, want %q", calledWith, "opencode")
	}
}

func TestAdapter_PostInject_AddsContextModePlugin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a config with no plugin entry.
	cfg := map[string]any{"mcp": map[string]any{}}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAdapter()
	err := a.PostInject(homeDir, []model.ComponentID{model.ComponentContextMode})
	if err != nil {
		t.Fatalf("PostInject error: %v", err)
	}

	// Read back and verify.
	result, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}

	plugins, ok := got["plugin"].([]any)
	if !ok {
		t.Fatal("expected plugin array in config")
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0] != "context-mode" {
		t.Fatalf("plugin[0] = %q, want %q", plugins[0], "context-mode")
	}
}

func TestAdapter_PostInject_Idempotent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a config that already has context-mode in plugin.
	cfg := map[string]any{
		"mcp":    map[string]any{},
		"plugin": []any{"context-mode"},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAdapter()
	err := a.PostInject(homeDir, []model.ComponentID{model.ComponentContextMode})
	if err != nil {
		t.Fatalf("PostInject error: %v", err)
	}

	// Read back — should still have exactly 1 entry.
	result, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}

	plugins := got["plugin"].([]any)
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin after idempotent call, got %d", len(plugins))
	}
}

func TestAdapter_PostInject_NoContextMode_NoOp(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := map[string]any{"mcp": map[string]any{}}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAdapter()
	// Pass components that don't include context-mode.
	err := a.PostInject(homeDir, []model.ComponentID{model.ComponentEngram})
	if err != nil {
		t.Fatalf("PostInject error: %v", err)
	}

	// Read back — should have no plugin key.
	result, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}

	if _, ok := got["plugin"]; ok {
		t.Fatal("plugin key should not exist when context-mode not selected")
	}
}

func TestAdapter_PostInject_ConfigNotExist_CreatesMinimal(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	a := NewAdapter()
	err := a.PostInject(homeDir, []model.ComponentID{model.ComponentContextMode})
	if err != nil {
		t.Fatalf("PostInject error: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "opencode.json")
	result, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}

	plugins, ok := got["plugin"].([]any)
	if !ok {
		t.Fatal("expected plugin array")
	}
	if len(plugins) != 1 || plugins[0] != "context-mode" {
		t.Fatalf("unexpected plugins: %v", plugins)
	}
}

func TestAdapter_PostInject_ExistingPlugins_Appends(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := map[string]any{
		"plugin": []any{"other-plugin"},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	a := NewAdapter()
	err := a.PostInject(homeDir, []model.ComponentID{model.ComponentContextMode})
	if err != nil {
		t.Fatalf("PostInject error: %v", err)
	}

	result, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}

	plugins := got["plugin"].([]any)
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if plugins[0] != "other-plugin" || plugins[1] != "context-mode" {
		t.Fatalf("unexpected plugins: %v", plugins)
	}
}
