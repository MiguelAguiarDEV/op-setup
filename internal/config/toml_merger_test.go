package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestTOMLMerger_EmptyFile_AddsServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "[mcp_servers.engram]") {
		t.Fatal("expected [mcp_servers.engram] section")
	}
	if !strings.Contains(content, `type = "local"`) {
		t.Fatal("expected type = local")
	}
	if !strings.Contains(content, `command = ["engram", "mcp", "--tools=agent"]`) {
		t.Fatalf("expected engram command, got:\n%s", content)
	}
}

func TestTOMLMerger_FileNotExist_CreatesIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.toml")

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
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

func TestTOMLMerger_ExistingContent_Preserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	existing := `# Codex config
model = "gpt-4"
temperature = 0.7
`
	os.WriteFile(path, []byte(existing), 0o644)

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, `model = "gpt-4"`) {
		t.Fatal("existing content should be preserved")
	}
	if !strings.Contains(content, "[mcp_servers.engram]") {
		t.Fatal("new section should be added")
	}
}

func TestTOMLMerger_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	m := NewTOMLMerger()
	servers := map[string]model.MCPServerConfig{
		"engram":   engramServer(),
		"context7": context7Server(),
	}

	// First merge.
	changed1, err := m.Merge(path, servers)
	if err != nil {
		t.Fatalf("first merge error: %v", err)
	}
	if !changed1 {
		t.Fatal("first merge should change")
	}

	// Second merge — same servers.
	changed2, err := m.Merge(path, servers)
	if err != nil {
		t.Fatalf("second merge error: %v", err)
	}
	if changed2 {
		t.Fatal("second merge should not change (idempotent)")
	}
}

func TestTOMLMerger_MultipleServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	m := NewTOMLMerger()
	servers := map[string]model.MCPServerConfig{
		"engram":   engramServer(),
		"context7": context7Server(),
		"github":   githubServer(),
	}

	changed, err := m.Merge(path, servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Sections should be sorted alphabetically.
	engIdx := strings.Index(content, "[mcp_servers.context7]")
	ghIdx := strings.Index(content, "[mcp_servers.engram]")
	gitIdx := strings.Index(content, "[mcp_servers.github]")

	if engIdx == -1 || ghIdx == -1 || gitIdx == -1 {
		t.Fatalf("missing sections in:\n%s", content)
	}
	if engIdx > ghIdx || ghIdx > gitIdx {
		t.Fatal("sections should be sorted alphabetically")
	}
}

func TestTOMLMerger_RemoteServer_HasURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"context7": context7Server(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, `url = "https://mcp.context7.com/mcp"`) {
		t.Fatalf("expected url field, got:\n%s", content)
	}
	if !strings.Contains(content, `type = "remote"`) {
		t.Fatal("expected type = remote")
	}
}

func TestTOMLMerger_GitHubServer_HasHeaders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	m := NewTOMLMerger()
	m.Merge(path, map[string]model.MCPServerConfig{
		"github": githubServer(),
	})

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, `headers.Authorization = "Bearer {env:GITHUB_MCP_PAT}"`) {
		t.Fatalf("expected Authorization header, got:\n%s", content)
	}
}

func TestTOMLMerger_EmptyServersMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte("# empty\n"), 0o644)

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false for empty servers map")
	}
}

func TestTOMLMerger_CommentedSectionHeader_Ignored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	existing := `# Example config
# [mcp_servers.fake] should not be treated as a section
model = "gpt-4"
`
	os.WriteFile(path, []byte(existing), 0o644)

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	// Original comment and content preserved.
	if !strings.Contains(content, "# [mcp_servers.fake]") {
		t.Fatal("commented section header should be preserved as-is")
	}
	if !strings.Contains(content, `model = "gpt-4"`) {
		t.Fatal("existing content should be preserved")
	}
	if !strings.Contains(content, "[mcp_servers.engram]") {
		t.Fatal("new section should be added")
	}
}

func TestTOMLMerger_CommentInsideSection_Preserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	existing := `[mcp_servers.engram]
# This server handles persistent memory
type = "local"
enabled = true
command = ["engram", "mcp", "--tools=agent"]
`
	os.WriteFile(path, []byte(existing), 0o644)

	m := NewTOMLMerger()
	// Re-merge same server — should be idempotent (comment is part of body).
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The comment inside the section makes the body differ from the rendered version,
	// so the section will be replaced (changed=true).
	if !changed {
		t.Fatal("expected changed=true (comment makes body differ)")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "[mcp_servers.engram]") {
		t.Fatal("section should exist after re-merge")
	}
}

func TestTOMLMerger_ExistingMCPSection_Preserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	existing := `[mcp_servers.existing]
type = "local"
enabled = true
command = ["existing-cmd"]
`
	os.WriteFile(path, []byte(existing), 0o644)

	m := NewTOMLMerger()
	changed, err := m.Merge(path, map[string]model.MCPServerConfig{
		"engram": engramServer(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "[mcp_servers.existing]") {
		t.Fatal("existing MCP section should be preserved")
	}
	if !strings.Contains(content, "[mcp_servers.engram]") {
		t.Fatal("new section should be added")
	}
}
