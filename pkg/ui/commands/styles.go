// Package commands provides CLI command-specific UI components.
package commands

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226"))

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))
)

// ColorSuccess returns text styled with the success color.
func ColorSuccess(text string) string {
	return SuccessStyle.Render(text)
}

// ColorWarning returns text styled with the warning color.
func ColorWarning(text string) string {
	return WarningStyle.Render(text)
}

// ColorInfo returns text styled with the info color.
func ColorInfo(text string) string {
	return InfoStyle.Render(text)
}
