package component

import (
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
