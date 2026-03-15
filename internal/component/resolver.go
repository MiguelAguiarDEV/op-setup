package component

import "github.com/MiguelAguiarDEV/op-setup/internal/model"

// Resolver determines how components map to specific agent configurations.
type Resolver struct{}

// NewResolver creates a new Resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve returns the MCPServerConfig for a component as it should appear
// in a specific agent's config. The config may vary per agent (e.g., different
// field names or formats).
func (r *Resolver) Resolve(comp Component, agent model.AgentID) model.MCPServerConfig {
	cfg := comp.Config

	// For Codex (TOML), remote servers use "url" field directly.
	// For JSON-based agents, the structure is the same.
	// No per-agent transformations needed in v1 — the merger handles format.
	_ = agent

	return cfg
}

// Compatible returns true if a component can be installed for a given agent.
// In v1, all components are compatible with all agents.
func (r *Resolver) Compatible(_ Component, _ model.AgentID) bool {
	return true
}

// ConfigKey returns the key name to use for a component in the agent's config.
// This is the key under the MCP servers map (e.g., "engram", "context-mode").
func (r *Resolver) ConfigKey(comp Component) string {
	return string(comp.ID)
}
