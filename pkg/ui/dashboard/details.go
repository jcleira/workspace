package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DetailsModel represents the right panel workspace details.
type DetailsModel struct {
	workspace *WorkspaceData
	statuses  []RepoStatus
	cursor    int
	width     int
	height    int
	focused   bool
	keys      KeyMap
	isLoading bool
}

// NewDetails creates a new details model.
func NewDetails(keys KeyMap) DetailsModel {
	return DetailsModel{
		keys:    keys,
		focused: false,
	}
}

// SetWorkspace sets the workspace to display.
func (m DetailsModel) SetWorkspace(ws *WorkspaceData) DetailsModel {
	m.workspace = ws
	m.cursor = 0
	return m
}

// SetStatuses sets the repo statuses.
func (m DetailsModel) SetStatuses(statuses []RepoStatus) DetailsModel {
	m.statuses = statuses
	m.isLoading = false
	if m.cursor >= len(statuses) && len(statuses) > 0 {
		m.cursor = len(statuses) - 1
	}
	return m
}

// SetLoading sets the loading state.
func (m DetailsModel) SetLoading(loading bool) DetailsModel {
	m.isLoading = loading
	return m
}

// SetSize sets the component dimensions.
func (m DetailsModel) SetSize(width, height int) DetailsModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets whether this panel is focused.
func (m DetailsModel) SetFocused(focused bool) DetailsModel {
	m.focused = focused
	return m
}

// SelectedRepo returns the currently selected repo index.
func (m DetailsModel) SelectedRepo() int {
	return m.cursor
}

// Init initializes the component.
func (m DetailsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m DetailsModel) Update(msg tea.Msg) (DetailsModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(keyMsg, m.keys.Down):
			if m.cursor < len(m.statuses)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the details panel.
func (m DetailsModel) View() string {
	var b strings.Builder

	if m.workspace == nil {
		b.WriteString(dimmedItemStyle.Render("Select a workspace"))
		return b.String()
	}

	name := strings.TrimPrefix(m.workspace.Name, "workspace-")
	b.WriteString(sectionHeaderStyle.Render("WORKSPACE DETAILS"))
	b.WriteString("\n")

	b.WriteString(titleStyle.Render(name))
	b.WriteString("\n")

	b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("Created: %s", m.workspace.Created.Format("Jan 2, 2006"))))
	b.WriteString("\n")
	b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("Path: %s", shortenPath(m.workspace.Path))))
	b.WriteString("\n\n")

	b.WriteString(sectionHeaderStyle.Render("REPOSITORIES"))
	b.WriteString("\n")

	if m.isLoading {
		b.WriteString(dimmedItemStyle.Render("Loading status..."))
		return b.String()
	}

	if len(m.statuses) == 0 {
		b.WriteString(dimmedItemStyle.Render("No repositories"))
		return b.String()
	}

	maxVisible := m.height - 8
	if maxVisible < 3 {
		maxVisible = 5
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.statuses) {
		end = len(m.statuses)
	}

	for i := start; i < end; i++ {
		status := m.statuses[i]
		b.WriteString(m.renderRepoStatus(status, i == m.cursor))
		b.WriteString("\n")
	}

	if len(m.statuses) > maxVisible {
		b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("... %d more repos", len(m.statuses)-maxVisible)))
	}

	return b.String()
}

func (m DetailsModel) renderRepoStatus(status RepoStatus, selected bool) string {
	var b strings.Builder

	cursor := "  "
	if selected && m.focused {
		cursor = cursorStyle.Render("> ")
	}

	nameStyle := normalItemStyle
	if selected && m.focused {
		nameStyle = selectedStyle
	}

	icon := Icon(status)
	statusText := StatusText(status)

	header := fmt.Sprintf("%s%s %s  %s", cursor, icon, nameStyle.Render(truncate(status.Name, 18)), statusText)
	b.WriteString(header)
	b.WriteString("\n")

	indent := "    "
	if status.Branch != "" {
		b.WriteString(indent)
		b.WriteString(treeStyle.Render("|-- "))
		b.WriteString(branchStyle.Render(status.Branch))
		b.WriteString("\n")
	}

	if status.HasUncommitted && status.UncommittedCount > 0 {
		b.WriteString(indent)
		b.WriteString(treeStyle.Render("|-- "))
		b.WriteString(statusModifiedStyle.Render(fmt.Sprintf("%d modified files", status.UncommittedCount)))
		b.WriteString("\n")
	}

	if status.UnpushedCount > 0 {
		b.WriteString(indent)
		b.WriteString(treeStyle.Render("|-- "))
		b.WriteString(statusModifiedStyle.Render(fmt.Sprintf("%d unpushed commits", status.UnpushedCount)))
		b.WriteString("\n")
	}

	if status.BehindCount > 0 {
		b.WriteString(indent)
		b.WriteString(treeStyle.Render("|-- "))
		b.WriteString(statusBehindStyle.Render(fmt.Sprintf("v %d behind remote", status.BehindCount)))
		b.WriteString("\n")
	}

	if !status.LastFetch.IsZero() {
		ago := formatTimeAgo(status.LastFetch)
		b.WriteString(indent)
		b.WriteString(treeStyle.Render("'-- "))
		b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("fetched %s", ago)))
		b.WriteString("\n")
	}

	return b.String()
}

func shortenPath(path string) string {
	home := "~"
	if strings.HasPrefix(path, "/Users/") {
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 4 {
			return home + "/" + parts[3]
		}
	}
	return path
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
