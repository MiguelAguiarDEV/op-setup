package adapter

import (
	"errors"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestNewAdapter_AllAgents(t *testing.T) {
	tests := []struct {
		agent    model.AgentID
		wantName string
	}{
		{model.AgentClaudeCode, "Claude Code"},
		{model.AgentOpenCode, "OpenCode"},
		{model.AgentCodex, "Codex"},
		{model.AgentGeminiCLI, "Gemini CLI"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agent), func(t *testing.T) {
			a, err := NewAdapter(tt.agent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if a.Name() != tt.wantName {
				t.Fatalf("Name() = %q, want %q", a.Name(), tt.wantName)
			}
			if a.Agent() != tt.agent {
				t.Fatalf("Agent() = %q, want %q", a.Agent(), tt.agent)
			}
		})
	}
}

func TestNewAdapter_UnknownAgent(t *testing.T) {
	_, err := NewAdapter("unknown-agent")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !errors.Is(err, ErrAgentNotSupported) {
		t.Fatalf("expected ErrAgentNotSupported, got %v", err)
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	r, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 4 {
		t.Fatalf("expected 4 adapters, got %d", r.Len())
	}

	// Verify all agents are registered.
	for _, agent := range model.AllAgents() {
		if _, ok := r.Get(agent); !ok {
			t.Fatalf("agent %q not found in default registry", agent)
		}
	}
}

func TestNewDefaultRegistry_AdapterProperties(t *testing.T) {
	r, _ := NewDefaultRegistry()

	tests := []struct {
		agent    model.AgentID
		strategy model.MCPStrategy
		key      string
	}{
		{model.AgentClaudeCode, model.StrategyMergeIntoJSON, "mcpServers"},
		{model.AgentOpenCode, model.StrategyMergeIntoJSON, "mcp"},
		{model.AgentCodex, model.StrategyMergeIntoTOML, "mcp_servers"},
		{model.AgentGeminiCLI, model.StrategyMergeIntoJSON, "mcpServers"},
	}

	for _, tt := range tests {
		t.Run(string(tt.agent), func(t *testing.T) {
			a, ok := r.Get(tt.agent)
			if !ok {
				t.Fatalf("agent %q not found", tt.agent)
			}
			if a.MCPStrategy() != tt.strategy {
				t.Fatalf("MCPStrategy() = %d, want %d", a.MCPStrategy(), tt.strategy)
			}
			if a.MCPConfigKey() != tt.key {
				t.Fatalf("MCPConfigKey() = %q, want %q", a.MCPConfigKey(), tt.key)
			}
		})
	}
}
