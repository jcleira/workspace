package dashboard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel represents the help overlay.
type HelpModel struct {
	width  int
	height int
	keys   KeyMap
}

// NewHelp creates a new help model.
func NewHelp(keys KeyMap) HelpModel {
	return HelpModel{
		keys: keys,
	}
}

// SetSize sets the component dimensions.
func (m HelpModel) SetSize(width, height int) HelpModel {
	m.width = width
	m.height = height
	return m
}

// Init initializes the component.
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	return m, nil
}

// View renders the help overlay.
func (m HelpModel) View() string {
	var b strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(60)

	helpTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(highlightColor).
		Width(15)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	b.WriteString(helpTitleStyle.Render("Workspace Dashboard Help"))
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine(keyStyle, descStyle, "j / k / up / down", "Move cursor up/down"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "h / l / left / right", "Switch focus between panels"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "/", "Filter workspaces"))
	b.WriteString("\n")

	b.WriteString(sectionStyle.Render("Actions"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine(keyStyle, descStyle, "Enter", "Switch to selected workspace"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "f", "Fetch all repos in workspace"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "p", "Pull all repos in workspace"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "g", "View diff for selected repo"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "G", "View staged diff for selected repo"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "r", "Refresh status"))
	b.WriteString("\n")

	b.WriteString(sectionStyle.Render("Workspace Management"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine(keyStyle, descStyle, "c", "Create new workspace"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "d", "Delete workspace"))
	b.WriteString("\n")

	b.WriteString(sectionStyle.Render("Other"))
	b.WriteString("\n")
	b.WriteString(renderHelpLine(keyStyle, descStyle, "?", "Toggle this help"))
	b.WriteString(renderHelpLine(keyStyle, descStyle, "q / Ctrl+C", "Quit"))

	content := boxStyle.Render(b.String())

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		)
	}

	return content
}

func renderHelpLine(keyStyle, descStyle lipgloss.Style, keyText, desc string) string {
	return keyStyle.Render(keyText) + descStyle.Render(desc) + "\n"
}
