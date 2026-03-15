package pipeline_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

func TestIntegration_FullRoundTrip(t *testing.T) {
	homeDir := t.TempDir()

	// Create config directories.
	claudeDir := filepath.Join(homeDir, ".claude")
	opencodeDir := filepath.Join(homeDir, ".config", "opencode")
	os.MkdirAll(claudeDir, 0o755)
	os.MkdirAll(opencodeDir, 0o755)

	// Write initial Claude config.
	claudeCfg := map[string]any{
		"permissions": map[string]any{"allow": []any{"read"}},
	}
	data, _ := json.MarshalIndent(claudeCfg, "", "  ")
	claudePath := filepath.Join(claudeDir, "settings.json")
	os.WriteFile(claudePath, data, 0o644)

	// Write initial OpenCode config.
	opencodeCfg := map[string]any{
		"theme": "dark",
	}
	data, _ = json.MarshalIndent(opencodeCfg, "", "  ")
	opencodePath := filepath.Join(opencodeDir, "opencode.json")
	os.WriteFile(opencodePath, data, 0o644)

	// Build registry and planner.
	registry, _ := adapter.NewDefaultRegistry()
	planner := &pipeline.Planner{
		Registry:   registry,
		HomeDir:    homeDir,
		BackupRoot: filepath.Join(homeDir, ".op-setup", "backups"),
	}

	// Plan for Claude + OpenCode with Engram + Context7.
	plan, err := planner.Plan(
		[]model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		[]model.ComponentID{model.ComponentEngram, model.ComponentContext7},
	)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}

	// Execute.
	orchestrator := pipeline.NewOrchestrator(nil)
	result := orchestrator.Execute(plan)

	if result.Err != nil {
		t.Fatalf("execution error: %v", result.Err)
	}
	if !result.Prepare.Success {
		t.Fatal("prepare should succeed")
	}
	if !result.Apply.Success {
		t.Fatal("apply should succeed")
	}

	// Verify Claude config.
	claudeData, _ := os.ReadFile(claudePath)
	var claudeResult map[string]any
	json.Unmarshal(claudeData, &claudeResult)

	// Original key preserved.
	if _, ok := claudeResult["permissions"]; !ok {
		t.Fatal("Claude: permissions key should be preserved")
	}

	// MCP servers added.
	mcpServers, ok := claudeResult["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("Claude: expected mcpServers key")
	}
	if _, ok := mcpServers["engram"]; !ok {
		t.Fatal("Claude: expected engram in mcpServers")
	}
	if _, ok := mcpServers["context7"]; !ok {
		t.Fatal("Claude: expected context7 in mcpServers")
	}

	// Verify OpenCode config.
	opencodeData, _ := os.ReadFile(opencodePath)
	var opencodeResult map[string]any
	json.Unmarshal(opencodeData, &opencodeResult)

	if opencodeResult["theme"] != "dark" {
		t.Fatal("OpenCode: theme key should be preserved")
	}

	mcp, ok := opencodeResult["mcp"].(map[string]any)
	if !ok {
		t.Fatal("OpenCode: expected mcp key")
	}
	if _, ok := mcp["engram"]; !ok {
		t.Fatal("OpenCode: expected engram in mcp")
	}

	// Verify backup was created.
	backupEntries, _ := os.ReadDir(filepath.Join(homeDir, ".op-setup", "backups"))
	if len(backupEntries) == 0 {
		t.Fatal("expected backup directory to be created")
	}
}

func TestIntegration_Idempotent(t *testing.T) {
	homeDir := t.TempDir()

	claudeDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(claudeDir, 0o755)

	registry, _ := adapter.NewDefaultRegistry()

	// First run.
	planner1 := &pipeline.Planner{
		Registry:   registry,
		HomeDir:    homeDir,
		BackupRoot: filepath.Join(homeDir, ".op-setup", "backups", "run1"),
	}
	plan1, _ := planner1.Plan(
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	orchestrator := pipeline.NewOrchestrator(nil)
	result1 := orchestrator.Execute(plan1)
	if result1.Err != nil {
		t.Fatalf("first run error: %v", result1.Err)
	}

	// Capture content after first run.
	data1, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))

	// Second run — same selections.
	planner2 := &pipeline.Planner{
		Registry:   registry,
		HomeDir:    homeDir,
		BackupRoot: filepath.Join(homeDir, ".op-setup", "backups", "run2"),
	}
	plan2, _ := planner2.Plan(
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	result2 := orchestrator.Execute(plan2)
	if result2.Err != nil {
		t.Fatalf("second run error: %v", result2.Err)
	}

	// Content should be identical (idempotent).
	data2, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if string(data1) != string(data2) {
		t.Fatalf("config changed on second run (not idempotent)\nfirst:\n%s\nsecond:\n%s", data1, data2)
	}
}
