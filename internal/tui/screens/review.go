package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// RenderReview renders the review screen before installation.
// profileName is the human-readable profile name (e.g. "Full Setup").
func RenderReview(profileName string, agents []string, components []string) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Review", "The following changes will be made"))
	b.WriteString("\n")

	b.WriteString(styles.SelectedStyle.Render("Profile:"))
	b.WriteString(" " + profileName)
	b.WriteString("\n\n")

	if len(agents) > 0 {
		b.WriteString(styles.SelectedStyle.Render("AI Tools:"))
		b.WriteString("\n")
		for _, a := range agents {
			b.WriteString(fmt.Sprintf("  %s %s\n", styles.Bullet, a))
		}
		b.WriteString("\n")
	}

	if len(components) > 0 {
		b.WriteString(styles.SelectedStyle.Render("MCP Servers:"))
		b.WriteString("\n")
		for _, c := range components {
			b.WriteString(fmt.Sprintf("  %s %s\n", styles.Bullet, c))
		}
		b.WriteString("\n")
	}

	// Show profile-specific info.
	if len(agents) == 0 && len(components) == 0 {
		b.WriteString(styles.UnselectedStyle.Render("Deploy agents, skills, scripts, and nvim config."))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.WarningStyle.Render("Existing configs will be backed up before modification."))
	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Press Enter to install • Esc to go back"))

	return b.String()
}
