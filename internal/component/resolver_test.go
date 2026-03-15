package component

import (
	"strings"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestResolver_Resolve_PreservesConfig(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentEngram)

	for _, agent := range model.AllAgents() {
		t.Run(string(agent), func(t *testing.T) {
			cfg := r.Resolve(comp, agent)
			if cfg.Type != model.MCPTypeLocal {
				t.Fatalf("Type = %q, want %q", cfg.Type, model.MCPTypeLocal)
			}
			if len(cfg.Command) != 3 {
				t.Fatalf("Command length = %d, want 3", len(cfg.Command))
			}
		})
	}
}

func TestResolver_Compatible_AllCombinations(t *testing.T) {
	r := NewResolver()
	components := All()
	agents := model.AllAgents()

	for _, comp := range components {
		for _, agent := range agents {
			if !r.Compatible(comp, agent) {
				t.Fatalf("component %q should be compatible with agent %q", comp.ID, agent)
			}
		}
	}
}

func TestResolver_ConfigKey(t *testing.T) {
	r := NewResolver()
	tests := []struct {
		id   model.ComponentID
		want string
	}{
		{model.ComponentEngram, "engram"},
		{model.ComponentContextMode, "context-mode"},
		{model.ComponentPlaywright, "playwright"},
		{model.ComponentGitHubMCP, "github"},
		{model.ComponentContext7, "context7"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			comp, _ := ByID(tt.id)
			got := r.ConfigKey(comp)
			if got != tt.want {
				t.Fatalf("ConfigKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolver_Resolve_RemoteComponent(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentGitHubMCP)

	cfg := r.Resolve(comp, model.AgentClaudeCode)
	if cfg.Type != model.MCPTypeRemote {
		t.Fatalf("Type = %q, want %q", cfg.Type, model.MCPTypeRemote)
	}
	if cfg.URL == "" {
		t.Fatal("URL should not be empty for remote component")
	}
	if cfg.Headers["Authorization"] == "" {
		t.Fatal("Authorization header should be set for GitHub MCP")
	}
}

// --- Env var resolution tests ---

func TestResolver_Resolve_ClaudeCode_KeepsEnvSyntax(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentGitHubMCP)

	cfg := r.Resolve(comp, model.AgentClaudeCode)
	auth := cfg.Headers["Authorization"]
	if !strings.Contains(auth, "{env:GITHUB_MCP_PAT}") {
		t.Fatalf("Claude Code should keep {env:X} syntax, got %q", auth)
	}
}

func TestResolver_Resolve_NonClaude_ResolvesEnvVars(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentGitHubMCP)

	t.Setenv("GITHUB_MCP_PAT", "test-token-123")

	for _, agent := range []model.AgentID{model.AgentOpenCode, model.AgentCodex, model.AgentGeminiCLI} {
		t.Run(string(agent), func(t *testing.T) {
			cfg := r.Resolve(comp, agent)
			auth := cfg.Headers["Authorization"]
			want := "Bearer test-token-123"
			if auth != want {
				t.Fatalf("Headers[Authorization] = %q, want %q", auth, want)
			}
		})
	}
}

func TestResolver_Resolve_NonClaude_EmptyEnvVar(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentGitHubMCP)

	t.Setenv("GITHUB_MCP_PAT", "")

	cfg := r.Resolve(comp, model.AgentOpenCode)
	auth := cfg.Headers["Authorization"]
	if auth != "Bearer " {
		t.Fatalf("Headers[Authorization] = %q, want %q", auth, "Bearer ")
	}
}

func TestResolver_Resolve_DoesNotMutateCatalog(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentGitHubMCP)

	t.Setenv("GITHUB_MCP_PAT", "mutated-value")

	// Resolve for non-Claude (triggers env var resolution).
	_ = r.Resolve(comp, model.AgentOpenCode)

	// Verify catalog is not mutated.
	original, _ := ByID(model.ComponentGitHubMCP)
	auth := original.Config.Headers["Authorization"]
	if !strings.Contains(auth, "{env:GITHUB_MCP_PAT}") {
		t.Fatalf("catalog was mutated: Headers[Authorization] = %q", auth)
	}
}

func TestResolver_Resolve_LocalComponent_NoEnvResolution(t *testing.T) {
	r := NewResolver()
	comp, _ := ByID(model.ComponentEngram)

	// Local components have no headers — should pass through unchanged.
	cfg := r.Resolve(comp, model.AgentOpenCode)
	if cfg.Type != model.MCPTypeLocal {
		t.Fatalf("Type = %q, want %q", cfg.Type, model.MCPTypeLocal)
	}
	if len(cfg.Headers) != 0 {
		t.Fatalf("local component should have no headers, got %d", len(cfg.Headers))
	}
}

func TestResolveEnvVars_MultiplePatterns(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Setenv("BAZ", "qux")

	result := resolveEnvVars("prefix-{env:FOO}-middle-{env:BAZ}-suffix")
	want := "prefix-bar-middle-qux-suffix"
	if result != want {
		t.Fatalf("resolveEnvVars() = %q, want %q", result, want)
	}
}

func TestResolveEnvVars_NoPattern(t *testing.T) {
	result := resolveEnvVars("no patterns here")
	if result != "no patterns here" {
		t.Fatalf("resolveEnvVars() = %q, want unchanged", result)
	}
}

func TestSupportsLazyEnv(t *testing.T) {
	if !supportsLazyEnv(model.AgentClaudeCode) {
		t.Fatal("Claude Code should support lazy env")
	}
	for _, agent := range []model.AgentID{model.AgentOpenCode, model.AgentCodex, model.AgentGeminiCLI} {
		if supportsLazyEnv(agent) {
			t.Fatalf("%q should NOT support lazy env", agent)
		}
	}
}
