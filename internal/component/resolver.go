package component

import (
	"os"
	"regexp"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// envVarPattern matches {env:VAR_NAME} placeholders.
var envVarPattern = regexp.MustCompile(`\{env:([^}]+)\}`)

// Resolver determines how components map to specific agent configurations.
type Resolver struct{}

// NewResolver creates a new Resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// supportsLazyEnv returns true if the agent natively resolves {env:X} at runtime.
func supportsLazyEnv(agent model.AgentID) bool {
	return agent == model.AgentClaudeCode
}

// resolveEnvVars replaces all {env:X} patterns with os.Getenv(X).
func resolveEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		sub := envVarPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		return os.Getenv(sub[1])
	})
}

// Resolve returns the MCPServerConfig for a component as it should appear
// in a specific agent's config.
//
// For agents that support {env:X} syntax natively (Claude Code), the config
// is returned as-is. For all other agents, {env:X} placeholders in headers
// and URLs are resolved to actual environment variable values.
func (r *Resolver) Resolve(comp Component, agent model.AgentID) model.MCPServerConfig {
	cfg := comp.Config

	if supportsLazyEnv(agent) {
		return cfg
	}

	// Resolve {env:X} in headers — create a new map to avoid mutating the catalog.
	if len(cfg.Headers) > 0 {
		resolved := make(map[string]string, len(cfg.Headers))
		for k, v := range cfg.Headers {
			resolved[k] = resolveEnvVars(v)
		}
		cfg.Headers = resolved
	}

	// Resolve {env:X} in URL if present.
	if strings.Contains(cfg.URL, "{env:") {
		cfg.URL = resolveEnvVars(cfg.URL)
	}

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
