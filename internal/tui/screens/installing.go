package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// RenderInstalling renders the installation progress screen.
func RenderInstalling(events []pipeline.ProgressEvent, total int) string {
	var b strings.Builder

	b.WriteString(RenderHeader("Installing", "Configuring MCP servers..."))
	b.WriteString("\n")

	for _, e := range events {
		var icon string
		switch e.Status {
		case pipeline.StatusRunning:
			icon = styles.WarningStyle.Render("⟳")
		case pipeline.StatusSucceeded:
			icon = styles.CheckMark
		case pipeline.StatusFailed:
			icon = styles.CrossMark
		case pipeline.StatusRolledBack:
			icon = styles.WarningStyle.Render("↩")
		default:
			icon = " "
		}

		line := fmt.Sprintf("  %s %s", icon, e.StepID)
		if e.Err != nil {
			line += " " + styles.ErrorStyle.Render("— "+e.Err.Error())
		}
		b.WriteString(line + "\n")
	}

	// Progress bar.
	completed := 0
	for _, e := range events {
		if e.Status == pipeline.StatusSucceeded || e.Status == pipeline.StatusFailed {
			completed++
		}
	}

	if total > 0 {
		b.WriteString("\n")
		b.WriteString(progressBar(completed, total, 40))
		b.WriteString("\n")
	}

	return b.String()
}

func progressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := styles.SelectedStyle.Render(strings.Repeat("█", filled))
	bar += styles.SubtitleStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("  %s %d/%d", bar, current, total)
}
