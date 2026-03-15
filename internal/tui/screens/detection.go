package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// DetectionItem represents a detected AI tool.
type DetectionItem struct {
	Name        string
	Agent       model.AgentID
	Installed   bool
	ConfigFound bool
}

// RenderDetection renders the detection results screen.
func RenderDetection(items []DetectionItem) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Detection Results", "Scanning for installed AI coding tools..."))
	b.WriteString("\n")

	for _, item := range items {
		var status string
		if item.Installed {
			status = styles.CheckMark + " " + styles.SelectedStyle.Render(item.Name)
			if item.ConfigFound {
				status += " " + styles.SubtitleStyle.Render("(config found)")
			} else {
				status += " " + styles.WarningStyle.Render("(no config file)")
			}
		} else {
			status = styles.CrossMark + " " + styles.DisabledStyle.Render(item.Name)
			status += " " + styles.SubtitleStyle.Render("(not installed)")
		}

		b.WriteString(fmt.Sprintf("  %s\n", status))
	}

	b.WriteString("\n")
	b.WriteString(RenderFooter("Press Enter to continue • Esc to go back"))

	return b.String()
}
