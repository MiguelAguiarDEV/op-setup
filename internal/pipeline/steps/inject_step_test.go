package steps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter/claude"
	oc "github.com/MiguelAguiarDEV/op-setup/internal/adapter/opencode"
	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestInjectStep_JSON_EmptyConfig(t *testing.T) {
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(cfgDir, 0o755)

	a := claude.NewAdapter()
	engram, _ := component.ByID(model.ComponentEngram)
	resolver := component.NewResolver()

	step := NewInjectStep(a, []component.Component{engram}, homeDir, resolver)

	if step.ID() != "inject-claude-code" {
		t.Fatalf("ID() = %q, want %q", step.ID(), "inject-claude-code")
	}

	if err := step.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if !step.Changed() {
		t.Fatal("expected changed=true for new config")
	}

	// Verify config was written.
	cfgPath := filepath.Join(cfgDir, "settings.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	var cfg map[string]any
	json.Unmarshal(data, &cfg)

	mcpServers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("expected mcpServers key")
	}
	if _, ok := mcpServers["engram"]; !ok {
		t.Fatal("expected engram in mcpServers")
	}
}

func TestInjectStep_JSON_Idempotent(t *testing.T) {
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(cfgDir, 0o755)

	a := claude.NewAdapter()
	engram, _ := component.ByID(model.ComponentEngram)
	resolver := component.NewResolver()

	// First inject.
	step1 := NewInjectStep(a, []component.Component{engram}, homeDir, resolver)
	step1.Run()

	// Second inject — should not change.
	step2 := NewInjectStep(a, []component.Component{engram}, homeDir, resolver)
	step2.Run()

	if step2.Changed() {
		t.Fatal("expected changed=false on second inject (idempotent)")
	}
}

func TestInjectStep_OpenCode_PostInject(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "") // Ensure default path resolution.
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".config", "opencode")
	os.MkdirAll(cfgDir, 0o755)

	a := oc.NewAdapter()
	ctxMode, _ := component.ByID(model.ComponentContextMode)
	resolver := component.NewResolver()

	step := NewInjectStep(a, []component.Component{ctxMode}, homeDir, resolver)
	if err := step.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	// Verify plugin entry was added.
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	data, _ := os.ReadFile(cfgPath)
	var cfg map[string]any
	json.Unmarshal(data, &cfg)

	plugins, ok := cfg["plugin"].([]any)
	if !ok {
		t.Fatal("expected plugin array")
	}
	found := false
	for _, p := range plugins {
		if p == "context-mode" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected context-mode in plugin array")
	}
}

func TestInjectStep_Rollback(t *testing.T) {
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(cfgDir, 0o755)

	cfgPath := filepath.Join(cfgDir, "settings.json")
	originalContent := `{"permissions": {"allow": ["read"]}}`
	os.WriteFile(cfgPath, []byte(originalContent), 0o644)

	// Create backup.
	snap := &backup.Snapshotter{Now: fixedTime}
	backupDir := filepath.Join(t.TempDir(), "backup")
	manifest, _ := snap.Create(backupDir, []string{cfgPath})

	// Inject.
	a := claude.NewAdapter()
	engram, _ := component.ByID(model.ComponentEngram)
	resolver := component.NewResolver()

	step := NewInjectStep(a, []component.Component{engram}, homeDir, resolver)
	step.SetManifest(&manifest)
	step.Run()

	// Verify config was modified.
	data, _ := os.ReadFile(cfgPath)
	if string(data) == originalContent {
		t.Fatal("config should have been modified")
	}

	// Rollback.
	if err := step.Rollback(); err != nil {
		t.Fatalf("Rollback error: %v", err)
	}

	// Verify original content restored.
	data, _ = os.ReadFile(cfgPath)
	if string(data) != originalContent {
		t.Fatalf("content after rollback = %q, want %q", data, originalContent)
	}
}

func TestInjectStep_Rollback_NoManifest(t *testing.T) {
	a := claude.NewAdapter()
	resolver := component.NewResolver()
	step := NewInjectStep(a, nil, "/tmp", resolver)

	err := step.Rollback()
	if err == nil {
		t.Fatal("expected error when no manifest")
	}
}

func TestInjectStep_MultipleComponents(t *testing.T) {
	homeDir := t.TempDir()
	cfgDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(cfgDir, 0o755)

	a := claude.NewAdapter()
	engram, _ := component.ByID(model.ComponentEngram)
	ctx7, _ := component.ByID(model.ComponentContext7)
	gh, _ := component.ByID(model.ComponentGitHubMCP)
	resolver := component.NewResolver()

	step := NewInjectStep(a, []component.Component{engram, ctx7, gh}, homeDir, resolver)
	if err := step.Run(); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "settings.json")
	data, _ := os.ReadFile(cfgPath)
	var cfg map[string]any
	json.Unmarshal(data, &cfg)

	mcpServers := cfg["mcpServers"].(map[string]any)
	if len(mcpServers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(mcpServers))
	}
}
