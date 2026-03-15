package screens

import (
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

const logo = `
                           _
  ___  _ __    ___  ___| |_ _   _ _ __
 / _ \| '_ \  / __|/ _ \ __| | | | '_ \
| (_) | |_) | \__ \  __/ |_| |_| | |_) |
 \___/| .__/  |___/\___|\__|\__,_| .__/
      |_|                         |_|
`

// RenderWelcome renders the welcome screen.
func RenderWelcome(version string) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render(logo))
	b.WriteString("\n")
	b.WriteString(styles.Tagline(version))
	b.WriteString("\n\n")
	b.WriteString(styles.UnselectedStyle.Render("Configure MCP servers for your AI coding tools."))
	b.WriteString("\n")
	b.WriteString(styles.UnselectedStyle.Render("Supports: Claude Code, OpenCode, Codex, Gemini CLI"))
	b.WriteString("\n\n")
	b.WriteString(RenderFooter("Press Enter to begin • q to quit"))

	return b.String()
}
