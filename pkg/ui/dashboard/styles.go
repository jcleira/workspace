package dashboard

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("62")
	successColor   = lipgloss.Color("46")
	warningColor   = lipgloss.Color("226")
	errorColor     = lipgloss.Color("196")
	infoColor      = lipgloss.Color("75")
	subtleColor    = lipgloss.Color("241")
	highlightColor = lipgloss.Color("39")

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtleColor).
			Padding(0, 1)

	focusedPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255"))

	// Footer styles
	footerBannerStyle = lipgloss.NewStyle().
				Border(lipgloss.Border{Top: "─"}).
				BorderForeground(subtleColor).
				Padding(0, 2).
				MarginTop(0)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	helpSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	// Status indicator styles
	statusCleanStyle = lipgloss.NewStyle().
				Foreground(successColor)

	statusModifiedStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	statusAheadStyle = lipgloss.NewStyle().
				Foreground(infoColor)

	statusBehindStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208"))

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(errorColor)

	// Selection styles
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor)

	cursorStyle = lipgloss.NewStyle().
			Foreground(highlightColor)

	// Item styles
	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimmedItemStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	// Section header style
	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				MarginBottom(1)

	// Config bar style
	configBarStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Padding(0, 1)

	// Branch style
	branchStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	// Tree indicator style
	treeStyle = lipgloss.NewStyle().
			Foreground(subtleColor)
)
