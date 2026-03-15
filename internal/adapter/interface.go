// Package adapter defines the interface for AI coding tool adapters
// and provides a registry for managing them.
package adapter

import "github.com/MiguelAguiarDEV/op-setup/internal/model"

// Adapter abstracts an AI coding tool's config management.
// Each supported tool (Claude Code, OpenCode, Codex, Gemini CLI) implements this.
type Adapter interface {
	// Name returns the human-readable name (e.g. "Claude Code").
	Name() string

	// Agent returns the canonical AgentID.
	Agent() model.AgentID

	// Detect checks whether the tool is installed and its config exists.
	// homeDir is the user's home directory (injectable for testing).
	Detect(homeDir string) (model.DetectResult, error)

	// ConfigPath returns the absolute path to the tool's config file.
	ConfigPath(homeDir string) string

	// MCPStrategy returns how MCP entries should be written.
	MCPStrategy() model.MCPStrategy

	// MCPConfigKey returns the JSON/TOML key under which MCP servers live
	// (e.g. "mcpServers", "mcp", "mcp_servers").
	MCPConfigKey() string

	// PostInject runs any extra actions after MCP config injection.
	// For example, OpenCode needs a "plugin" array entry for context-mode.
	PostInject(homeDir string, components []model.ComponentID) error
}
