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
			styles.SuccessStyle.Render("Setup Complete!"),
			"All stages completed successfully.",
		))
	} else {
		b.WriteString(RenderHeader(
			styles.ErrorStyle.Render("Setup Failed"),
			result.Err.Error(),
		))
	}

	b.WriteString("\n")

	// Show Install stage results.
	if len(result.Install.Steps) > 0 {
		b.WriteString(styles.SelectedStyle.Render("Tools Installed:"))
		b.WriteString("\n")
		renderStepResults(&b, result.Install.Steps)
		b.WriteString("\n")
	}

	// Show Deploy stage results.
	if len(result.Deploy.Steps) > 0 {
		b.WriteString(styles.SelectedStyle.Render("Dotfiles Deployed:"))
		b.WriteString("\n")
		renderStepResults(&b, result.Deploy.Steps)
		b.WriteString("\n")
	}

	// Show Apply stage results.
	if len(result.Apply.Steps) > 0 {
		b.WriteString(styles.SelectedStyle.Render("MCP Servers Configured:"))
		b.WriteString("\n")
		renderStepResults(&b, result.Apply.Steps)
		b.WriteString("\n")
	}

	// Show rollback info if applicable.
	if result.Rollback != nil {
		b.WriteString(styles.WarningStyle.Render("Rollback:"))
		b.WriteString("\n")
		if len(result.Rollback.Steps) > 0 {
			renderStepResults(&b, result.Rollback.Steps)
		} else {
			b.WriteString("  Original configs restored.\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(RenderFooter("Press q to quit"))

	return b.String()
}

// renderStepResults writes step results to the builder.
func renderStepResults(b *strings.Builder, steps []pipeline.StepResult) {
	for _, sr := range steps {
		var icon string
		switch sr.Status {
		case pipeline.StatusSucceeded, pipeline.StatusRolledBack:
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
