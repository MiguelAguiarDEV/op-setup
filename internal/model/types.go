// Package model defines the core domain types shared across all packages.
package model

import "fmt"

// AgentID identifies an AI coding tool.
type AgentID string

const (
	AgentClaudeCode AgentID = "claude-code"
	AgentOpenCode   AgentID = "opencode"
	AgentCodex      AgentID = "codex"
	AgentGeminiCLI  AgentID = "gemini-cli"
)

// AllAgents returns every supported AgentID in display order.
func AllAgents() []AgentID {
	return []AgentID{
		AgentClaudeCode,
		AgentOpenCode,
		AgentCodex,
		AgentGeminiCLI,
	}
}

// ComponentID identifies an MCP server component.
type ComponentID string

const (
	ComponentEngram      ComponentID = "engram"
	ComponentContextMode ComponentID = "context-mode"
	ComponentPlaywright  ComponentID = "playwright"
	ComponentGitHubMCP   ComponentID = "github"
	ComponentContext7    ComponentID = "context7"
)

// AllComponents returns every supported ComponentID in display order.
func AllComponents() []ComponentID {
	return []ComponentID{
		ComponentEngram,
		ComponentContextMode,
		ComponentPlaywright,
		ComponentGitHubMCP,
		ComponentContext7,
	}
}

// MCPStrategy defines how MCP configs are written to an agent's config file.
type MCPStrategy int

const (
	// StrategyMergeIntoJSON merges MCP entries into a JSON config file
	// under a specific key (e.g. "mcpServers" or "mcp").
	StrategyMergeIntoJSON MCPStrategy = iota

	// StrategyMergeIntoTOML merges MCP entries into a TOML config file
	// under [mcp_servers.X] table sections.
	StrategyMergeIntoTOML
)

// MCPType distinguishes local command-based servers from remote URL-based ones.
type MCPType string

const (
	MCPTypeLocal  MCPType = "local"
	MCPTypeRemote MCPType = "remote"
)

// MCPServerConfig is the canonical representation of one MCP server entry.
type MCPServerConfig struct {
	Type    MCPType           `json:"type"`
	Command []string          `json:"command,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Enabled bool              `json:"enabled"`
}

// InstallerID identifies an installable tool.
type InstallerID string

const (
	InstallerOpenCode    InstallerID = "opencode"
	InstallerEngram      InstallerID = "engram"
	InstallerContextMode InstallerID = "context-mode"
	InstallerPlaywright  InstallerID = "playwright"
)

// AllInstallers returns every supported InstallerID in display order.
func AllInstallers() []InstallerID {
	return []InstallerID{
		InstallerOpenCode,
		InstallerEngram,
		InstallerContextMode,
		InstallerPlaywright,
	}
}

// SetupProfile determines which stages of the pipeline are active.
type SetupProfile string

const (
	// ProfileFull runs all stages: install tools, deploy dotfiles, configure MCP.
	ProfileFull SetupProfile = "full"

	// ProfileMCPOnly only configures MCP servers (original v1 behavior).
	ProfileMCPOnly SetupProfile = "mcp-only"

	// ProfileDotfilesOnly only deploys dotfiles (agents, skills, nvim config).
	ProfileDotfilesOnly SetupProfile = "dotfiles-only"
)

// AllProfiles returns every supported SetupProfile in display order.
func AllProfiles() []SetupProfile {
	return []SetupProfile{
		ProfileFull,
		ProfileMCPOnly,
		ProfileDotfilesOnly,
	}
}

// ProfileName returns a human-readable name for the profile.
func (p SetupProfile) Name() string {
	switch p {
	case ProfileFull:
		return "Full Setup"
	case ProfileMCPOnly:
		return "MCP Servers Only"
	case ProfileDotfilesOnly:
		return "Dotfiles Only"
	default:
		return string(p)
	}
}

// ProfileDescription returns a description of what the profile does.
func (p SetupProfile) Description() string {
	switch p {
	case ProfileFull:
		return "Install tools, deploy dotfiles, and configure MCP servers"
	case ProfileMCPOnly:
		return "Only configure MCP servers in AI tool config files"
	case ProfileDotfilesOnly:
		return "Only deploy agents, skills, scripts, and nvim config"
	default:
		return ""
	}
}

// ParseProfile validates and returns a SetupProfile from a string.
// Returns an error if the string is not a valid profile.
func ParseProfile(s string) (SetupProfile, error) {
	switch SetupProfile(s) {
	case ProfileFull, ProfileMCPOnly, ProfileDotfilesOnly:
		return SetupProfile(s), nil
	default:
		return "", fmt.Errorf("invalid profile %q: must be one of: full, mcp-only, dotfiles-only", s)
	}
}

// DetectResult holds the outcome of detecting an AI coding tool.
type DetectResult struct {
	// Installed is true if the tool's binary was found in PATH.
	Installed bool

	// BinaryPath is the resolved path to the binary (empty if not found).
	BinaryPath string

	// ConfigPath is the expected config file path.
	ConfigPath string

	// ConfigFound is true if the config file exists on disk.
	ConfigFound bool
}
