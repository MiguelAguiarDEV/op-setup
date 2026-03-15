package screens

import "strings"

// ProfileItem represents a setup profile in the selection screen.
type ProfileItem struct {
	Name        string
	Description string
}

// RenderProfile renders the profile selection screen.
func RenderProfile(items []ProfileItem, cursor int) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Select Setup Profile", "Choose what to configure"))
	b.WriteString("\n")

	selectItems := make([]SingleSelectItem, len(items))
	for i, item := range items {
		selectItems[i] = SingleSelectItem{
			Label:       item.Name,
			Description: item.Description,
		}
	}

	b.WriteString(RenderSingleSelect(selectItems, cursor))
	b.WriteString("\n")
	b.WriteString(RenderFooter("↑/↓ to navigate • Enter to select • Esc to go back"))

	return b.String()
}
