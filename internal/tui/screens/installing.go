package screens

import (
	"fmt"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui/styles"
)

// installingSubtitle returns a profile-aware subtitle for the installing screen.
func installingSubtitle(profile model.SetupProfile) string {
	switch profile {
	case model.ProfileFull:
		return "Installing tools, deploying dotfiles, configuring MCP servers..."
	case model.ProfileMCPOnly:
		return "Configuring MCP servers..."
	case model.ProfileDotfilesOnly:
		return "Deploying dotfiles..."
	default:
		return "Setting up..."
	}
}

// RenderInstalling renders the installation progress screen.
func RenderInstalling(profile model.SetupProfile, events []pipeline.ProgressEvent, total int, dryRun bool) string {
	var b strings.Builder

	title := "Installing"
	if dryRun {
		title = "[DRY RUN] Installing"
	}
	b.WriteString(RenderHeader(title, installingSubtitle(profile)))
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
