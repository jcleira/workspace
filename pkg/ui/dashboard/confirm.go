package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel represents a confirmation dialog.
type ConfirmModel struct {
	action    ConfirmAction
	width     int
	height    int
	keys      KeyMap
	confirmed bool
	canceled  bool
	deleting  bool
}

// NewConfirm creates a new confirmation dialog.
func NewConfirm(action ConfirmAction, keys KeyMap) ConfirmModel {
	return ConfirmModel{
		action: action,
		keys:   keys,
	}
}

// SetSize sets the component dimensions.
func (m ConfirmModel) SetSize(width, height int) ConfirmModel {
	m.width = width
	m.height = height
	return m
}

// IsConfirmed returns true if the user confirmed.
func (m ConfirmModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCanceled returns true if the user canceled.
func (m ConfirmModel) IsCanceled() bool {
	return m.canceled
}

// IsDone returns true if the dialog is complete.
func (m ConfirmModel) IsDone() bool {
	if m.deleting {
		return false
	}
	return m.confirmed || m.canceled
}

// SetDeleting sets the deleting state.
func (m ConfirmModel) SetDeleting(deleting bool) ConfirmModel {
	m.deleting = deleting
	return m
}

// IsDeleting returns true if a deletion is in progress.
func (m ConfirmModel) IsDeleting() bool {
	return m.deleting
}

// Init initializes the component.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	if m.deleting {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Confirm):
			m.confirmed = true
			return m, nil
		case key.Matches(keyMsg, m.keys.Cancel):
			m.canceled = true
			return m, nil
		}
	}

	return m, nil
}

// View renders the confirmation dialog.
func (m ConfirmModel) View() string {
	var b strings.Builder

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(warningColor).
		Padding(1, 2).
		Width(50)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(warningColor)

	if m.deleting {
		b.WriteString(titleStyle.Render("Deleting Workspace"))
		b.WriteString("\n\n")
		b.WriteString("Deleting workspace and cleaning up branches...")
		b.WriteString("\n\n")
		if m.action.Workspace != "" {
			fmt.Fprintf(&b, "Workspace: %s\n", m.action.Workspace)
		}
		b.WriteString("\n")
		b.WriteString("Please wait...")

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

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginTop(1).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(highlightColor)

	title := fmt.Sprintf("Confirm %s", m.action.Action.String())
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	b.WriteString(descStyle.Render(m.action.Description))
	b.WriteString("\n\n")

	if m.action.Workspace != "" {
		fmt.Fprintf(&b, "Workspace: %s\n", m.action.Workspace)
	}
	if m.action.Repo != "" {
		fmt.Fprintf(&b, "Repository: %s\n", m.action.Repo)
	}

	b.WriteString("\n")
	b.WriteString(keyStyle.Render("[y]") + " Yes  ")
	b.WriteString(keyStyle.Render("[n/Esc]") + " No")

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
