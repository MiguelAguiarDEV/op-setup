// Package component defines the MCP server components that op-setup can install.
package component

import (
	"os"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Component describes an installable MCP server.
type Component struct {
	// ID is the canonical identifier.
	ID model.ComponentID

	// Name is the human-readable display name.
	Name string

	// Description is a short summary of what this component does.
	Description string

	// Config is the canonical MCP server configuration.
	Config model.MCPServerConfig

	// EnvVars lists environment variables required for this component.
	// Empty if none are needed.
	EnvVars []string
}

var catalog = []Component{
	{
		ID:          model.ComponentEngram,
		Name:        "Engram",
		Description: "Persistent cross-session memory",
		Config: model.MCPServerConfig{
			Type:    model.MCPTypeLocal,
			Command: []string{"engram", "mcp", "--tools=agent"},
			Enabled: true,
		},
	},
	{
		ID:          model.ComponentContextMode,
		Name:        "Context Mode",
		Description: "Context optimization — reduces usage by ~98%",
		Config: model.MCPServerConfig{
			Type:    model.MCPTypeLocal,
			Command: []string{"context-mode"},
			Enabled: true,
		},
	},
	{
		ID:          model.ComponentPlaywright,
		Name:        "Playwright",
		Description: "Browser automation via MCP",
		Config: model.MCPServerConfig{
			Type:    model.MCPTypeLocal,
			Command: []string{"npx", "-y", "@executeautomation/playwright-mcp-server"},
			Enabled: true,
		},
	},
	{
		ID:          model.ComponentGitHubMCP,
		Name:        "GitHub MCP",
		Description: "GitHub API access via Copilot MCP",
		Config: model.MCPServerConfig{
			Type:    model.MCPTypeRemote,
			URL:     "https://api.githubcopilot.com/mcp",
			Headers: map[string]string{"Authorization": "Bearer {env:GITHUB_MCP_PAT}"},
			Enabled: true,
		},
		EnvVars: []string{"GITHUB_MCP_PAT"},
	},
	{
		ID:          model.ComponentContext7,
		Name:        "Context7",
		Description: "Up-to-date framework documentation",
		Config: model.MCPServerConfig{
			Type:    model.MCPTypeRemote,
			URL:     "https://mcp.context7.com/mcp",
			Enabled: true,
		},
	},
}

// All returns the full catalog of available components.
func All() []Component {
	out := make([]Component, len(catalog))
	copy(out, catalog)
	return out
}

// EnvSatisfied returns true if all required env vars for the component are set.
func EnvSatisfied(c Component) bool {
	for _, env := range c.EnvVars {
		if os.Getenv(env) == "" {
			return false
		}
	}
	return true
}

// ByID returns a component by its ID. Returns false if not found.
func ByID(id model.ComponentID) (Component, bool) {
	for _, c := range catalog {
		if c.ID == id {
			return c, true
		}
	}
	return Component{}, false
}
