package screens

import (
	"os"
	"strings"
)

// ComponentItem represents a component in the selection screen.
type ComponentItem struct {
	Name     string
	Desc     string
	Selected bool
	EnvVars  []string // Required env vars.
}

// RenderComponents renders the component selection screen.
func RenderComponents(items []ComponentItem, cursor int) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Select MCP Servers", "Choose which MCP servers to install"))
	b.WriteString("\n")

	selectItems := make([]SelectItem, len(items))
	for i, item := range items {
		si := SelectItem{
			Label:       item.Name,
			Description: item.Desc,
			Selected:    item.Selected,
		}

		// Check for missing env vars.
		for _, env := range item.EnvVars {
			if os.Getenv(env) == "" {
				si.DisabledMsg = env + " not set"
				break
			}
		}

		selectItems[i] = si
	}

	b.WriteString(RenderMultiSelect(selectItems, cursor))
	b.WriteString("\n")
	b.WriteString(RenderFooter("Space to toggle • Enter to continue • Esc to go back"))

	return b.String()
}
