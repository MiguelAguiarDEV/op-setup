package component

import (
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestAll_Returns5Components(t *testing.T) {
	all := All()
	if len(all) != 5 {
		t.Fatalf("expected 5 components, got %d", len(all))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	a := All()
	a[0].Name = "mutated"
	b := All()
	if b[0].Name == "mutated" {
		t.Fatal("All() should return a copy, not a reference to internal state")
	}
}

func TestAll_NoDuplicateIDs(t *testing.T) {
	all := All()
	seen := make(map[model.ComponentID]bool, len(all))
	for _, c := range all {
		if seen[c.ID] {
			t.Fatalf("duplicate ComponentID: %q", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestByID_AllKnown(t *testing.T) {
	tests := []struct {
		id       model.ComponentID
		wantName string
	}{
		{model.ComponentEngram, "Engram"},
		{model.ComponentContextMode, "Context Mode"},
		{model.ComponentPlaywright, "Playwright"},
		{model.ComponentGitHubMCP, "GitHub MCP"},
		{model.ComponentContext7, "Context7"},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			c, ok := ByID(tt.id)
			if !ok {
				t.Fatalf("component %q not found", tt.id)
			}
			if c.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", c.Name, tt.wantName)
			}
		})
	}
}

func TestByID_Unknown(t *testing.T) {
	_, ok := ByID("nonexistent")
	if ok {
		t.Fatal("expected false for unknown ComponentID")
	}
}

func TestEngram_Config(t *testing.T) {
	c, ok := ByID(model.ComponentEngram)
	if !ok {
		t.Fatal("engram not found")
	}
	if c.Config.Type != model.MCPTypeLocal {
		t.Fatalf("Type = %q, want %q", c.Config.Type, model.MCPTypeLocal)
	}
	wantCmd := []string{"engram", "mcp", "--tools=agent"}
	if len(c.Config.Command) != len(wantCmd) {
		t.Fatalf("Command length = %d, want %d", len(c.Config.Command), len(wantCmd))
	}
	for i, v := range c.Config.Command {
		if v != wantCmd[i] {
			t.Fatalf("Command[%d] = %q, want %q", i, v, wantCmd[i])
		}
	}
}

func TestContextMode_Config(t *testing.T) {
	c, ok := ByID(model.ComponentContextMode)
	if !ok {
		t.Fatal("context-mode not found")
	}
	if c.Config.Type != model.MCPTypeLocal {
		t.Fatalf("Type = %q, want %q", c.Config.Type, model.MCPTypeLocal)
	}
	if len(c.Config.Command) != 1 || c.Config.Command[0] != "context-mode" {
		t.Fatalf("Command = %v, want [context-mode]", c.Config.Command)
	}
}

func TestPlaywright_Config(t *testing.T) {
	c, ok := ByID(model.ComponentPlaywright)
	if !ok {
		t.Fatal("playwright not found")
	}
	if c.Config.Type != model.MCPTypeLocal {
		t.Fatalf("Type = %q, want %q", c.Config.Type, model.MCPTypeLocal)
	}
	wantCmd := []string{"npx", "-y", "@executeautomation/playwright-mcp-server"}
	if len(c.Config.Command) != len(wantCmd) {
		t.Fatalf("Command length = %d, want %d", len(c.Config.Command), len(wantCmd))
	}
	for i, v := range c.Config.Command {
		if v != wantCmd[i] {
			t.Fatalf("Command[%d] = %q, want %q", i, v, wantCmd[i])
		}
	}
}

func TestGitHubMCP_Config(t *testing.T) {
	c, ok := ByID(model.ComponentGitHubMCP)
	if !ok {
		t.Fatal("github not found")
	}
	if c.Config.Type != model.MCPTypeRemote {
		t.Fatalf("Type = %q, want %q", c.Config.Type, model.MCPTypeRemote)
	}
	if c.Config.URL != "https://api.githubcopilot.com/mcp" {
		t.Fatalf("URL = %q, want %q", c.Config.URL, "https://api.githubcopilot.com/mcp")
	}
	if len(c.EnvVars) != 1 || c.EnvVars[0] != "GITHUB_MCP_PAT" {
		t.Fatalf("EnvVars = %v, want [GITHUB_MCP_PAT]", c.EnvVars)
	}
	if c.Config.Headers["Authorization"] != "Bearer {env:GITHUB_MCP_PAT}" {
		t.Fatalf("Authorization header = %q, want %q",
			c.Config.Headers["Authorization"], "Bearer {env:GITHUB_MCP_PAT}")
	}
}

func TestContext7_Config(t *testing.T) {
	c, ok := ByID(model.ComponentContext7)
	if !ok {
		t.Fatal("context7 not found")
	}
	if c.Config.Type != model.MCPTypeRemote {
		t.Fatalf("Type = %q, want %q", c.Config.Type, model.MCPTypeRemote)
	}
	if c.Config.URL != "https://mcp.context7.com/mcp" {
		t.Fatalf("URL = %q, want %q", c.Config.URL, "https://mcp.context7.com/mcp")
	}
	if len(c.EnvVars) != 0 {
		t.Fatalf("EnvVars = %v, want empty", c.EnvVars)
	}
}

// --- EnvSatisfied ---

func TestEnvSatisfied_NoEnvVars(t *testing.T) {
	c := Component{ID: "test", EnvVars: nil}
	if !EnvSatisfied(c) {
		t.Fatal("component with no env vars should be satisfied")
	}
}

func TestEnvSatisfied_AllSet(t *testing.T) {
	t.Setenv("TEST_VAR_A", "value-a")
	t.Setenv("TEST_VAR_B", "value-b")
	c := Component{ID: "test", EnvVars: []string{"TEST_VAR_A", "TEST_VAR_B"}}
	if !EnvSatisfied(c) {
		t.Fatal("component with all env vars set should be satisfied")
	}
}

func TestEnvSatisfied_OneMissing(t *testing.T) {
	t.Setenv("TEST_VAR_A", "value-a")
	t.Setenv("TEST_VAR_B", "")
	c := Component{ID: "test", EnvVars: []string{"TEST_VAR_A", "TEST_VAR_B"}}
	if EnvSatisfied(c) {
		t.Fatal("component with missing env var should NOT be satisfied")
	}
}

func TestEnvSatisfied_EmptyValue(t *testing.T) {
	t.Setenv("TEST_VAR_EMPTY", "")
	c := Component{ID: "test", EnvVars: []string{"TEST_VAR_EMPTY"}}
	if EnvSatisfied(c) {
		t.Fatal("component with empty env var should NOT be satisfied")
	}
}

func TestAll_EnabledByDefault(t *testing.T) {
	for _, c := range All() {
		if !c.Config.Enabled {
			t.Fatalf("component %q should be enabled by default", c.ID)
		}
	}
}
