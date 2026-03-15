// Package styles provides the Kanagawa-inspired color theme and lipgloss styles.
package styles

import "github.com/charmbracelet/lipgloss"

// Kanagawa color palette.
var (
	ColorBase    = lipgloss.Color("#1F1F28")
	ColorSurface = lipgloss.Color("#2A2A37")
	ColorText    = lipgloss.Color("#DCD7BA")
	ColorSubtext = lipgloss.Color("#727169")
	ColorBlue    = lipgloss.Color("#7E9CD8")
	ColorGreen   = lipgloss.Color("#76946A")
	ColorRed     = lipgloss.Color("#C34043")
	ColorYellow  = lipgloss.Color("#DCA561")
	ColorPurple  = lipgloss.Color("#957FB8")
	ColorCyan    = lipgloss.Color("#7FB4CA")
)

// Cursor is the selection indicator.
const Cursor = "▸ "

// Pre-built styles.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			MarginBottom(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	DisabledStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Strikethrough(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			MarginTop(1)

	CheckMark = SuccessStyle.Render("✓")
	CrossMark = ErrorStyle.Render("✗")
	Bullet    = lipgloss.NewStyle().Foreground(ColorPurple).Render("●")
)

// Tagline returns the version tagline.
func Tagline(version string) string {
	return SubtitleStyle.Render("op-setup " + version + " — AI coding environment configurator")
}
