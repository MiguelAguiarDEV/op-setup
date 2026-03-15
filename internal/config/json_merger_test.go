package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func engramServer() model.MCPServerConfig {
	return model.MCPServerConfig{
		Type:    model.MCPTypeLocal,
		Command: []string{"engram", "mcp", "--tools=agent"},
		Enabled: true,
	}
}

func context7Server() model.MCPServerConfig {
	return model.MCPServerConfig{
		Type:    model.MCPTypeRemote,
		URL:     "https://mcp.context7.com/mcp",
		Enabled: true,
	}
}

func githubServer() model.MCPServerConfig {
	return model.MCPServerConfig{
		Type:    model.MCPTypeRemote,
		URL:     "https://api.githubcopilot.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer {env:GITHUB_MCP_PAT}"},
		Enabled: true,
	}
}

func TestJSONMerger_EmptyFile_AddsServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true for new file")
	}

	// Verify file was created and is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	mcpServers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("expected mcpServers key")
	}
	if _, ok := mcpServers["engram"]; !ok {
		t.Fatal("expected engram entry in mcpServers")
	}
}

func TestJSONMerger_FileNotExist_CreatesIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "settings.json")

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestJSONMerger_ExistingConfig_PreservesKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write existing config with other keys.
	existing := map[string]any{
		"permissions": map[string]any{"allow": []any{"read"}},
		"theme":       "dark",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(path, data, 0o644)

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	// Verify existing keys preserved.
	result, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(result, &cfg)

	if cfg["theme"] != "dark" {
		t.Fatalf("theme = %v, want %q", cfg["theme"], "dark")
	}
	if _, ok := cfg["permissions"]; !ok {
		t.Fatal("permissions key should be preserved")
	}
	if _, ok := cfg["mcpServers"]; !ok {
		t.Fatal("mcpServers key should be added")
	}
}

func TestJSONMerger_ExistingMCPServers_MergesNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write config with existing MCP server.
	existing := map[string]any{
		"mcpServers": map[string]any{
			"existing": map[string]any{
				"type":    "local",
				"command": []any{"existing-cmd"},
				"enabled": true,
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(path, data, 0o644)

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	result, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(result, &cfg)

	mcpServers := cfg["mcpServers"].(map[string]any)
	if _, ok := mcpServers["existing"]; !ok {
		t.Fatal("existing MCP server should be preserved")
	}
	if _, ok := mcpServers["engram"]; !ok {
		t.Fatal("engram should be added")
	}
}

func TestJSONMerger_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	m := NewJSONMerger()
	servers := map[string]model.MCPServerConfig{
		"engram":   engramServer(),
		"context7": context7Server(),
	}

	// First merge.
	changed1, err := m.Merge(path, "mcpServers", servers)
	if err != nil {
		t.Fatalf("first merge error: %v", err)
	}
	if !changed1 {
		t.Fatal("first merge should change")
	}

	// Capture file content after first merge.
	data1, _ := os.ReadFile(path)

	// Second merge — same servers.
	changed2, err := m.Merge(path, "mcpServers", servers)
	if err != nil {
		t.Fatalf("second merge error: %v", err)
	}
	if changed2 {
		t.Fatal("second merge should not change (idempotent)")
	}

	// File should be unchanged.
	data2, _ := os.ReadFile(path)
	if !bytes.Equal(data1, data2) {
		t.Fatal("file content should not change on idempotent merge")
	}
}

func TestJSONMerger_InvalidJSON_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json{{{"), 0o644)

	m := NewJSONMerger()
	_, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestJSONMerger_MultipleServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	m := NewJSONMerger()
	servers := map[string]model.MCPServerConfig{
		"engram":   engramServer(),
		"context7": context7Server(),
		"github":   githubServer(),
	}

	changed, err := m.Merge(path, "mcpServers", servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	result, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(result, &cfg)

	mcpServers := cfg["mcpServers"].(map[string]any)
	if len(mcpServers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(mcpServers))
	}

	// Verify GitHub MCP has headers.
	gh := mcpServers["github"].(map[string]any)
	headers := gh["headers"].(map[string]any)
	if headers["Authorization"] != "Bearer {env:GITHUB_MCP_PAT}" {
		t.Fatalf("Authorization = %v", headers["Authorization"])
	}
}

func TestJSONMerger_DifferentKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcp", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	result, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(result, &cfg)

	if _, ok := cfg["mcp"]; !ok {
		t.Fatal("expected 'mcp' key (not 'mcpServers')")
	}
}

func TestJSONMerger_UpdateExistingServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Write config with an old version of engram.
	existing := map[string]any{
		"mcpServers": map[string]any{
			"engram": map[string]any{
				"type":    "local",
				"command": []any{"old-engram"},
				"enabled": true,
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(path, data, 0o644)

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(), // New command
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true when updating existing server")
	}

	result, _ := os.ReadFile(path)
	var cfg map[string]any
	json.Unmarshal(result, &cfg)

	engram := cfg["mcpServers"].(map[string]any)["engram"].(map[string]any)
	cmd := engram["command"].([]any)
	if len(cmd) != 3 || cmd[0] != "engram" {
		t.Fatalf("command should be updated, got %v", cmd)
	}
}

func TestJSONMerger_EmptyServersMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	os.WriteFile(path, []byte("{}"), 0o644)

	m := NewJSONMerger()
	changed, err := m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false for empty servers map")
	}
}

func TestJSONMerger_TrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	m := NewJSONMerger()
	m.Merge(path, "mcpServers", map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})

	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatal("file should not be empty")
	}
	if data[len(data)-1] != '\n' {
		t.Fatal("file should end with newline")
	}
}
