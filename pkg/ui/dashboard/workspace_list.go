package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// WorkspaceListModel represents the left panel workspace list.
type WorkspaceListModel struct {
	workspaces   []WorkspaceData
	statuses     map[string][]RepoStatus
	cursor       int
	scrollOffset int
	width        int
	height       int
	focused      bool
	filterMode   bool
	filterText   string
	filteredIdxs []int
	keys         KeyMap
}

// NewWorkspaceList creates a new workspace list model.
func NewWorkspaceList(workspaces []WorkspaceData, keys KeyMap) WorkspaceListModel {
	m := WorkspaceListModel{
		workspaces:   workspaces,
		statuses:     make(map[string][]RepoStatus),
		keys:         keys,
		focused:      true,
		filteredIdxs: make([]int, len(workspaces)),
	}
	for i := range workspaces {
		m.filteredIdxs[i] = i
	}
	return m
}

// SetWorkspaces updates the workspace list.
func (m WorkspaceListModel) SetWorkspaces(workspaces []WorkspaceData) WorkspaceListModel {
	m.workspaces = workspaces
	m.resetFilter()
	if m.cursor >= len(m.filteredIdxs) && len(m.filteredIdxs) > 0 {
		m.cursor = len(m.filteredIdxs) - 1
	}
	m.scrollOffset = 0
	return m
}

// SetStatuses updates the status map.
func (m WorkspaceListModel) SetStatuses(statuses map[string][]RepoStatus) WorkspaceListModel {
	m.statuses = statuses
	return m
}

// SetSize sets the component dimensions.
func (m WorkspaceListModel) SetSize(width, height int) WorkspaceListModel {
	m.width = width
	m.height = height
	return m
}

// SetFocused sets whether this panel is focused.
func (m WorkspaceListModel) SetFocused(focused bool) WorkspaceListModel {
	m.focused = focused
	return m
}

// SelectedIndex returns the actual index of the selected workspace.
func (m WorkspaceListModel) SelectedIndex() int {
	if len(m.filteredIdxs) == 0 {
		return -1
	}
	return m.filteredIdxs[m.cursor]
}

// SelectedWorkspace returns the selected workspace or nil.
func (m WorkspaceListModel) SelectedWorkspace() *WorkspaceData {
	idx := m.SelectedIndex()
	if idx < 0 || idx >= len(m.workspaces) {
		return nil
	}
	return &m.workspaces[idx]
}

func (m *WorkspaceListModel) resetFilter() {
	m.filteredIdxs = make([]int, len(m.workspaces))
	for i := range m.workspaces {
		m.filteredIdxs[i] = i
	}
	m.filterText = ""
	m.filterMode = false
	m.scrollOffset = 0
}

func (m *WorkspaceListModel) applyFilter() {
	if m.filterText == "" {
		m.resetFilter()
		return
	}

	m.filteredIdxs = make([]int, 0)
	search := strings.ToLower(m.filterText)
	for i, ws := range m.workspaces {
		if strings.Contains(strings.ToLower(ws.Name), search) {
			m.filteredIdxs = append(m.filteredIdxs, i)
		}
	}

	if m.cursor >= len(m.filteredIdxs) && len(m.filteredIdxs) > 0 {
		m.cursor = len(m.filteredIdxs) - 1
	}
	m.scrollOffset = 0
}

// Init initializes the component.
func (m WorkspaceListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m WorkspaceListModel) Update(msg tea.Msg) (WorkspaceListModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.filterMode {
		if keyMsg.Type == tea.KeyEscape {
			m.resetFilter()
			return m, nil
		}
		if keyMsg.Type == tea.KeyEnter {
			m.filterMode = false
			return m, nil
		}
		if keyMsg.Type == tea.KeyBackspace {
			if m.filterText != "" {
				m.filterText = m.filterText[:len(m.filterText)-1]
				m.applyFilter()
			}
			return m, nil
		}
		if keyMsg.Type == tea.KeyRunes {
			m.filterText += string(keyMsg.Runes)
			m.applyFilter()
		}
		return m, nil
	}

	maxVisible := m.maxVisibleItems()

	switch {
	case key.Matches(keyMsg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}
	case key.Matches(keyMsg, m.keys.Down):
		if m.cursor < len(m.filteredIdxs)-1 {
			m.cursor++
			if m.cursor >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.cursor - maxVisible + 1
			}
		}
	case key.Matches(keyMsg, m.keys.Filter):
		m.filterMode = true
		m.filterText = ""
	}

	return m, nil
}

func (m WorkspaceListModel) maxVisibleItems() int {
	maxVisible := m.height - 4
	if m.scrollOffset > 0 {
		maxVisible--
	}
	if maxVisible < 1 {
		maxVisible = 5
	}
	return maxVisible
}

// View renders the workspace list.
func (m WorkspaceListModel) View() string {
	var b strings.Builder

	header := "WORKSPACES"
	if m.filterMode {
		header = fmt.Sprintf("FILTER: %s_", m.filterText)
	} else if m.filterText != "" {
		header = fmt.Sprintf("WORKSPACES (%d/%d)", len(m.filteredIdxs), len(m.workspaces))
	}
	b.WriteString(sectionHeaderStyle.Render(header))
	b.WriteString("\n")

	if len(m.filteredIdxs) == 0 {
		b.WriteString(dimmedItemStyle.Render("No workspaces found"))
		return b.String()
	}

	maxVisible := m.maxVisibleItems()

	start := m.scrollOffset
	end := start + maxVisible
	if end > len(m.filteredIdxs) {
		end = len(m.filteredIdxs)
	}

	if start > 0 {
		b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("  ... %d above", start)))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		idx := m.filteredIdxs[i]
		ws := m.workspaces[idx]

		cursor := "  "
		if i == m.cursor && m.focused {
			cursor = cursorStyle.Render("> ")
		}

		name := strings.TrimPrefix(ws.Name, "workspace-")

		statusIndicator := " "
		if statuses, ok := m.statuses[ws.Path]; ok {
			issues := CountWorkspaceIssues(statuses)
			if issues > 0 {
				statusIndicator = statusModifiedStyle.Render(fmt.Sprintf("%d", issues))
			} else {
				statusIndicator = statusCleanStyle.Render("✓")
			}
		}

		nameStyle := normalItemStyle
		if i == m.cursor && m.focused {
			nameStyle = selectedStyle
		}

		line := fmt.Sprintf("%s%s %s", cursor, statusIndicator, nameStyle.Render(truncate(name, 35)))
		b.WriteString(line)
		b.WriteString("\n")
	}

	remaining := len(m.filteredIdxs) - end
	if remaining > 0 {
		b.WriteString(dimmedItemStyle.Render(fmt.Sprintf("  ... %d more", remaining)))
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
