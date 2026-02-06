package setup

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("62")
	successColor   = lipgloss.Color("46")
	warningColor   = lipgloss.Color("226")
	errorColor     = lipgloss.Color("196")
	subtleColor    = lipgloss.Color("241")
	highlightColor = lipgloss.Color("39")

	wizardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Width(65)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			MarginBottom(1)

	stepIndicatorStyle = lipgloss.NewStyle().
				Foreground(subtleColor).
				MarginBottom(1)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtleColor).
			Padding(0, 1).
			Width(55)

	focusedInputBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(highlightColor).
				Padding(0, 1).
				Width(55)

	validationSuccessStyle = lipgloss.NewStyle().
				Foreground(successColor)

	validationWarningStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	validationErrorStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	selectedPathStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true)

	pickerItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	pickerSelectedStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true)

	pickerDirStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	summaryLabelStyle = lipgloss.NewStyle().
				Foreground(subtleColor)

	summaryValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true)

	successTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor).
				MarginBottom(1)
)
