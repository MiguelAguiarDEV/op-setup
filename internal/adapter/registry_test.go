package adapter

import (
	"errors"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// stubAdapter is a minimal Adapter implementation for testing the registry.
type stubAdapter struct {
	name     string
	agent    model.AgentID
	strategy model.MCPStrategy
	key      string
}

func (s *stubAdapter) Name() string                   { return s.name }
func (s *stubAdapter) Agent() model.AgentID           { return s.agent }
func (s *stubAdapter) MCPStrategy() model.MCPStrategy { return s.strategy }
func (s *stubAdapter) MCPConfigKey() string           { return s.key }

func (s *stubAdapter) Detect(string) (model.DetectResult, error) {
	return model.DetectResult{}, nil
}

func (s *stubAdapter) ConfigPath(string) string { return "" }

func (s *stubAdapter) PostInject(string, []model.ComponentID) error {
	return nil
}

func newStub(name string, agent model.AgentID) *stubAdapter {
	return &stubAdapter{name: name, agent: agent, strategy: model.StrategyMergeIntoJSON, key: "mcpServers"}
}

func TestNewRegistry_Empty(t *testing.T) {
	r, err := NewRegistry()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 0 {
		t.Fatalf("expected 0 adapters, got %d", r.Len())
	}
}

func TestNewRegistry_WithAdapters(t *testing.T) {
	a1 := newStub("Claude Code", model.AgentClaudeCode)
	a2 := newStub("OpenCode", model.AgentOpenCode)

	r, err := NewRegistry(a1, a2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Len() != 2 {
		t.Fatalf("expected 2 adapters, got %d", r.Len())
	}
}

func TestNewRegistry_DuplicateReturnsError(t *testing.T) {
	a1 := newStub("Claude Code", model.AgentClaudeCode)
	a2 := newStub("Claude Code Dup", model.AgentClaudeCode)

	_, err := NewRegistry(a1, a2)
	if err == nil {
		t.Fatal("expected error for duplicate adapter")
	}
	if !errors.Is(err, ErrDuplicateAdapter) {
		t.Fatalf("expected ErrDuplicateAdapter, got %v", err)
	}
}

func TestRegistry_Register_Duplicate(t *testing.T) {
	r, _ := NewRegistry()
	a := newStub("Claude Code", model.AgentClaudeCode)

	if err := r.Register(a); err != nil {
		t.Fatalf("first register should succeed: %v", err)
	}
	if err := r.Register(a); err == nil {
		t.Fatal("second register should fail")
	}
}

func TestRegistry_Get_Found(t *testing.T) {
	a := newStub("Claude Code", model.AgentClaudeCode)
	r, _ := NewRegistry(a)

	got, ok := r.Get(model.AgentClaudeCode)
	if !ok {
		t.Fatal("expected adapter to be found")
	}
	if got.Name() != "Claude Code" {
		t.Fatalf("got name %q, want %q", got.Name(), "Claude Code")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r, _ := NewRegistry()

	_, ok := r.Get(model.AgentClaudeCode)
	if ok {
		t.Fatal("expected adapter to not be found")
	}
}

func TestRegistry_All_Sorted(t *testing.T) {
	// Register in non-alphabetical order
	a1 := newStub("OpenCode", model.AgentOpenCode)
	a2 := newStub("Claude Code", model.AgentClaudeCode)
	a3 := newStub("Gemini CLI", model.AgentGeminiCLI)
	a4 := newStub("Codex", model.AgentCodex)

	r, err := NewRegistry(a1, a2, a3, a4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := r.All()
	if len(all) != 4 {
		t.Fatalf("expected 4 adapters, got %d", len(all))
	}

	// Should be sorted by AgentID: claude-code, codex, gemini-cli, opencode
	expected := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentCodex,
		model.AgentGeminiCLI,
		model.AgentOpenCode,
	}
	for i, a := range all {
		if a.Agent() != expected[i] {
			t.Fatalf("All()[%d].Agent() = %q, want %q", i, a.Agent(), expected[i])
		}
	}
}

func TestRegistry_All_Empty(t *testing.T) {
	r, _ := NewRegistry()
	all := r.All()
	if len(all) != 0 {
		t.Fatalf("expected 0 adapters, got %d", len(all))
	}
}
