package screens

import "strings"

// AgentItem represents an agent in the selection screen.
type AgentItem struct {
	Name     string
	Detected bool
	Selected bool
}

// RenderAgents renders the agent selection screen.
func RenderAgents(items []AgentItem, cursor int) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Select AI Tools", "Choose which tools to configure with MCP servers"))
	b.WriteString("\n")

	selectItems := make([]SelectItem, len(items))
	for i, item := range items {
		si := SelectItem{
			Label:    item.Name,
			Selected: item.Selected,
		}
		if !item.Detected {
			si.Disabled = true
			si.DisabledMsg = "not installed"
		}
		selectItems[i] = si
	}

	b.WriteString(RenderMultiSelect(selectItems, cursor))
	b.WriteString("\n")
	b.WriteString(RenderFooter("Space to toggle • Enter to continue • Esc to go back"))

	return b.String()
}
