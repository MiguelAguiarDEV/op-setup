package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// RenderComplete renders the completion screen.
func RenderComplete(result pipeline.ExecutionResult) string {
	var b strings.Builder

	if result.Err == nil {
		b.WriteString(RenderHeader(
			styles.SuccessStyle.Render("Installation Complete!"),
			"All MCP servers have been configured successfully.",
		))
	} else {
		b.WriteString(RenderHeader(
			styles.ErrorStyle.Render("Installation Failed"),
			result.Err.Error(),
		))
	}

	b.WriteString("\n")

	// Show step results.
	if len(result.Apply.Steps) > 0 {
		b.WriteString(styles.SelectedStyle.Render("Results:"))
		b.WriteString("\n")
		for _, sr := range result.Apply.Steps {
			var icon string
			switch sr.Status {
			case pipeline.StatusSucceeded:
				icon = styles.CheckMark
			case pipeline.StatusFailed:
				icon = styles.CrossMark
			default:
				icon = " "
			}
			line := fmt.Sprintf("  %s %s", icon, sr.StepID)
			if sr.Err != nil {
				line += " " + styles.ErrorStyle.Render("— "+sr.Err.Error())
			}
			b.WriteString(line + "\n")
		}
	}

	// Show rollback info if applicable.
	if result.Rollback != nil {
		b.WriteString("\n")
		b.WriteString(styles.WarningStyle.Render("Rollback was executed — original configs restored."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderFooter("Press q to quit"))

	return b.String()
}
