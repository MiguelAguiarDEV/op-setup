// Package screens provides render functions for each TUI screen.
package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// SelectItem represents an item in a multi-select list.
type SelectItem struct {
	Label       string
	Description string
	Selected    bool
	Disabled    bool
	DisabledMsg string
}

// RenderHeader renders a screen title and subtitle.
func RenderHeader(title, subtitle string) string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render(title))
	b.WriteString("\n")
	if subtitle != "" {
		b.WriteString(styles.SubtitleStyle.Render(subtitle))
		b.WriteString("\n")
	}
	return b.String()
}

// RenderFooter renders help text at the bottom.
func RenderFooter(help string) string {
	return styles.HelpStyle.Render(help)
}

// RenderMultiSelect renders a multi-select list with cursor.
func RenderMultiSelect(items []SelectItem, cursor int) string {
	var b strings.Builder

	for i, item := range items {
		prefix := "  "
		if i == cursor {
			prefix = styles.Cursor
		}

		checkbox := "[ ]"
		if item.Selected {
			checkbox = "[" + styles.CheckMark + "]"
		}

		label := item.Label
		if item.Disabled {
			label = styles.DisabledStyle.Render(label)
			if item.DisabledMsg != "" {
				label += " " + styles.WarningStyle.Render("("+item.DisabledMsg+")")
			}
		} else if item.Selected {
			label = styles.SelectedStyle.Render(label)
		}

		line := fmt.Sprintf("%s%s %s", prefix, checkbox, label)
		if item.Description != "" && !item.Disabled {
			line += " " + styles.SubtitleStyle.Render("— "+item.Description)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
