package adapter

import (
	"github.com/MiguelAguiarDEV/op-setup/internal/model"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter/claude"
	"github.com/MiguelAguiarDEV/op-setup/internal/adapter/codex"
	"github.com/MiguelAguiarDEV/op-setup/internal/adapter/gemini"
	"github.com/MiguelAguiarDEV/op-setup/internal/adapter/opencode"
)

// Compile-time interface checks.
var (
	_ Adapter = (*claude.Adapter)(nil)
	_ Adapter = (*opencode.Adapter)(nil)
	_ Adapter = (*codex.Adapter)(nil)
	_ Adapter = (*gemini.Adapter)(nil)
)

// NewAdapter creates a single adapter for the given AgentID.
// Returns AgentNotSupportedError for unknown agents.
func NewAdapter(agent model.AgentID) (Adapter, error) {
	switch agent {
	case model.AgentClaudeCode:
		return claude.NewAdapter(), nil
	case model.AgentOpenCode:
		return opencode.NewAdapter(), nil
	case model.AgentCodex:
		return codex.NewAdapter(), nil
	case model.AgentGeminiCLI:
		return gemini.NewAdapter(), nil
	default:
		return nil, &AgentNotSupportedError{Agent: agent}
	}
}

// NewDefaultRegistry creates a Registry pre-populated with all supported adapters.
func NewDefaultRegistry() (*Registry, error) {
	return NewRegistry(
		claude.NewAdapter(),
		opencode.NewAdapter(),
		codex.NewAdapter(),
		gemini.NewAdapter(),
	)
}
