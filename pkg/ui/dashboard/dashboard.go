package dashboard

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jcleira/workspace/pkg/config"
	"github.com/jcleira/workspace/pkg/workspace"
)

const refreshInterval = 10 * time.Second

// Pane represents which panel is focused.
type Pane int

const (
	PaneWorkspaceList Pane = iota
	PaneDetails
)

// DashboardModel is the main dashboard model.
type DashboardModel struct {
	workspaces    []WorkspaceData
	selectedIndex int
	repoStatuses  map[string][]RepoStatus

	focusedPane  Pane
	showHelp     bool
	showConfirm  bool
	showDiff     bool
	confirmModel ConfirmModel
	diffView     DiffViewModel

	workspaceList WorkspaceListModel
	details       DetailsModel
	help          HelpModel

	workspaceManager *workspace.Manager
	workspaceService *workspace.Service
	configManager    *config.ConfigManager

	width  int
	height int
	ready  bool

	refreshing    bool
	lastRefresh   time.Time
	keys          KeyMap
	statusMessage string

	selectedPath string
	quitting     bool
	openSetup    bool
}

// NewDashboard creates a new dashboard model.
func NewDashboard(wm *workspace.Manager, cm *config.ConfigManager) DashboardModel {
	keys := DefaultKeyMap()
	return DashboardModel{
		repoStatuses:     make(map[string][]RepoStatus),
		keys:             keys,
		workspaceList:    NewWorkspaceList(nil, keys),
		details:          NewDetails(keys),
		help:             NewHelp(keys),
		diffView:         NewDiffView(keys),
		workspaceManager: wm,
		workspaceService: workspace.NewService(wm),
		configManager:    cm,
		focusedPane:      PaneWorkspaceList,
	}
}

// Init initializes the dashboard.
func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadWorkspaces(),
		m.tickCmd(),
	)
}

func (m DashboardModel) tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// StatusClearedMsg signals that status message should be cleared.
type StatusClearedMsg struct{}

func (m DashboardModel) clearStatusAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return StatusClearedMsg{}
	})
}

func (m DashboardModel) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		workspaces, err := m.workspaceManager.GetWorkspaces()
		if err != nil {
			return ErrorMsg{Error: err, Context: "loading workspaces"}
		}
		return WorkspacesLoadedMsg{Workspaces: WorkspacesToData(workspaces)}
	}
}

func (m DashboardModel) refreshSelectedStatus() tea.Cmd {
	if len(m.workspaces) == 0 {
		return nil
	}

	if m.selectedIndex >= len(m.workspaces) {
		return nil
	}

	ws := m.workspaces[m.selectedIndex]

	return func() tea.Msg {
		statuses := FetchWorkspaceStatus(ws.Path, ws.Projects)
		return WorkspaceStatusMsg{
			WorkspacePath: ws.Path,
			Statuses:      statuses,
		}
	}
}

func (m DashboardModel) refreshAllStatuses() []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.workspaces))
	for _, ws := range m.workspaces {
		cmds = append(cmds, func() tea.Msg {
			statuses := FetchWorkspaceStatus(ws.Path, ws.Projects)
			return WorkspaceStatusMsg{
				WorkspacePath: ws.Path,
				Statuses:      statuses,
			}
		})
	}
	return cmds
}

// Update handles all messages.
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m = m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		if m.showHelp {
			if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Cancel) {
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}

		if m.showConfirm {
			var cmd tea.Cmd
			m.confirmModel, cmd = m.confirmModel.Update(msg)
			if m.confirmModel.IsConfirmed() && !m.confirmModel.IsDeleting() {
				m.confirmModel = m.confirmModel.SetDeleting(true)
				cmd = m.executeConfirmedAction()
				return m, cmd
			}
			if m.confirmModel.IsCanceled() {
				m.showConfirm = false
			}
			return m, cmd
		}

		if m.showDiff {
			if key.Matches(msg, m.keys.Cancel) || key.Matches(msg, m.keys.Quit) {
				m.showDiff = false
				return m, nil
			}
			var cmd tea.Cmd
			m.diffView, cmd = m.diffView.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Settings):
			m.openSetup = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = true
			return m, nil

		case key.Matches(msg, m.keys.Left):
			m.focusedPane = PaneWorkspaceList
			m.workspaceList = m.workspaceList.SetFocused(true)
			m.details = m.details.SetFocused(false)
			return m, nil

		case key.Matches(msg, m.keys.Right):
			m.focusedPane = PaneDetails
			m.workspaceList = m.workspaceList.SetFocused(false)
			m.details = m.details.SetFocused(true)
			return m, nil

		case key.Matches(msg, m.keys.Select):
			switch m.focusedPane {
			case PaneWorkspaceList:
				ws := m.workspaceList.SelectedWorkspace()
				if ws != nil {
					m.selectedPath = ws.Path
					m.quitting = true
					return m, tea.Quit
				}
			case PaneDetails:
				return m.handleDiffAction(false)
			}
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.refreshing = true
			m.statusMessage = "Refreshing..."
			return m, tea.Batch(m.refreshSelectedStatus(), m.clearStatusAfterDelay())

		case key.Matches(msg, m.keys.Fetch):
			return m.handleFetchAction()

		case key.Matches(msg, m.keys.Pull):
			return m.handlePullAction()

		case key.Matches(msg, m.keys.Delete):
			return m.handleDeleteAction()

		case key.Matches(msg, m.keys.Diff):
			return m.handleDiffAction(false)

		case key.Matches(msg, m.keys.DiffStaged):
			return m.handleDiffAction(true)
		}

	case TickMsg:
		if !m.refreshing && time.Since(m.lastRefresh) > refreshInterval/2 {
			return m, tea.Batch(m.refreshSelectedStatus(), m.tickCmd())
		}
		return m, m.tickCmd()

	case WorkspacesLoadedMsg:
		m.workspaces = msg.Workspaces
		m.selectedIndex = 0
		m.workspaceList = m.workspaceList.SetWorkspaces(msg.Workspaces)
		if len(msg.Workspaces) > 0 {
			ws := m.workspaces[0]
			m.details = m.details.SetWorkspace(&ws)
			m.details = m.details.SetLoading(true)
			cmds = append(cmds, m.refreshAllStatuses()...)
		}
		return m, tea.Batch(cmds...)

	case WorkspaceStatusMsg:
		m.repoStatuses[msg.WorkspacePath] = msg.Statuses
		m.workspaceList = m.workspaceList.SetStatuses(m.repoStatuses)
		m.lastRefresh = time.Now()
		m.refreshing = false

		if m.selectedIndex < len(m.workspaces) &&
			m.workspaces[m.selectedIndex].Path == msg.WorkspacePath {
			m.details = m.details.SetStatuses(msg.Statuses)
		}
		return m, nil

	case ActionCompleteMsg:
		if msg.Action == ActionDelete {
			m.showConfirm = false
		}
		switch {
		case msg.Output != "":
			m.statusMessage = msg.Output
		case msg.Error != nil:
			m.statusMessage = msg.Error.Error()
		default:
			m.statusMessage = ""
		}
		return m, tea.Batch(m.loadWorkspaces(), m.refreshSelectedStatus())

	case ErrorMsg:
		m.statusMessage = msg.Error.Error()
		return m, nil

	case StatusClearedMsg:
		if m.statusMessage == "Refreshing..." {
			m.statusMessage = ""
		}
		return m, nil

	case DiffLoadedMsg:
		m.diffView = m.diffView.SetLoading(false)
		if msg.Error != nil {
			m.diffView = m.diffView.SetError(msg.Error)
		} else {
			m.diffView = m.diffView.SetContent(msg.Content, msg.RepoName, msg.Staged)
		}
		return m, nil
	}

	if m.focusedPane == PaneWorkspaceList {
		var cmd tea.Cmd
		m.workspaceList, cmd = m.workspaceList.Update(msg)
		cmds = append(cmds, cmd)

		newIdx := m.workspaceList.SelectedIndex()
		if newIdx != m.selectedIndex && newIdx >= 0 && newIdx < len(m.workspaces) {
			m.selectedIndex = newIdx
			ws := m.workspaces[m.selectedIndex]
			m.details = m.details.SetWorkspace(&ws)

			if statuses, ok := m.repoStatuses[ws.Path]; ok {
				m.details = m.details.SetStatuses(statuses)
			} else {
				m.details = m.details.SetLoading(true)
				cmds = append(cmds, m.refreshSelectedStatus())
			}
		}
	} else {
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m DashboardModel) updateLayout() DashboardModel {
	leftWidth := m.width * 2 / 5
	if leftWidth < 25 {
		leftWidth = 25
	}
	if leftWidth > 50 {
		leftWidth = 50
	}
	rightWidth := m.width - leftWidth - 4

	contentHeight := m.height - 6
	innerHeight := contentHeight - 2

	m.workspaceList = m.workspaceList.SetSize(leftWidth, innerHeight)
	m.details = m.details.SetSize(rightWidth, innerHeight)
	m.help = m.help.SetSize(m.width, m.height)
	if m.showConfirm {
		m.confirmModel = m.confirmModel.SetSize(m.width, m.height)
	}
	if m.showDiff {
		m.diffView = m.diffView.SetSize(m.width, m.height)
	}

	return m
}

func (m DashboardModel) handleFetchAction() (tea.Model, tea.Cmd) {
	ws := m.workspaceList.SelectedWorkspace()
	if ws == nil {
		return m, nil
	}

	return m, executeAction(ActionFetch, ws.Path, ws.Projects, "")
}

func (m DashboardModel) handlePullAction() (tea.Model, tea.Cmd) {
	ws := m.workspaceList.SelectedWorkspace()
	if ws == nil {
		return m, nil
	}

	return m, executeAction(ActionPull, ws.Path, ws.Projects, "")
}

func (m DashboardModel) handleDeleteAction() (tea.Model, tea.Cmd) {
	ws := m.workspaceList.SelectedWorkspace()
	if ws == nil {
		return m, nil
	}

	if ws.Name == workspace.DefaultWorkspaceName || ws.Name == "default" {
		return m, nil
	}

	m.confirmModel = NewConfirm(ConfirmAction{
		Action:      ActionDelete,
		Workspace:   ws.Name,
		Description: "This will delete the workspace and all associated branches.",
	}, m.keys)
	m.confirmModel = m.confirmModel.SetSize(m.width, m.height)
	m.showConfirm = true

	return m, nil
}

func (m DashboardModel) handleDiffAction(staged bool) (tea.Model, tea.Cmd) {
	ws := m.workspaceList.SelectedWorkspace()
	if ws == nil {
		return m, nil
	}

	if m.focusedPane != PaneDetails || len(ws.Projects) == 0 {
		return m, nil
	}

	repoIdx := m.details.SelectedRepo()
	if repoIdx < 0 || repoIdx >= len(ws.Projects) {
		return m, nil
	}

	repoPath := filepath.Join(ws.Path, ws.Projects[repoIdx])
	repoName := ws.Projects[repoIdx]

	m.showDiff = true
	m.diffView = m.diffView.SetSize(m.width, m.height)
	m.diffView = m.diffView.SetLoading(true)

	return m, loadDiffContent(repoPath, repoName, staged)
}

func (m DashboardModel) executeConfirmedAction() tea.Cmd {
	action := m.confirmModel.action

	switch action.Action {
	case ActionDelete:
		return func() tea.Msg {
			name := strings.TrimPrefix(action.Workspace, "workspace-")
			input := workspace.DeleteInput{
				Name:           name,
				DeleteBranches: true,
			}
			output, err := m.workspaceService.Delete(input)
			if err != nil {
				return ActionCompleteMsg{
					Action:    ActionDelete,
					Workspace: action.Workspace,
					Success:   false,
					Error:     err,
				}
			}

			if len(output.SkippedBranches) > 0 {
				skippedMsg := fmt.Sprintf("%d branch(es) not deleted due to unpushed commits", len(output.SkippedBranches))
				return ActionCompleteMsg{
					Action:    ActionDelete,
					Workspace: action.Workspace,
					Success:   true,
					Output:    skippedMsg,
				}
			}

			return ActionCompleteMsg{
				Action:    ActionDelete,
				Workspace: action.Workspace,
				Success:   true,
			}
		}
	case ActionFetch, ActionPull, ActionCreate, ActionSwitch, ActionDiff:
		return nil
	}

	return nil
}

// View renders the dashboard.
func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	if !m.ready {
		return "Loading..."
	}

	if m.showHelp {
		return m.help.View()
	}

	if m.showConfirm {
		return m.confirmModel.View()
	}

	if m.showDiff {
		return m.diffView.View()
	}

	header := m.renderHeader()
	configBar := m.renderConfigBar()
	footer := m.renderFooter()
	content := m.renderContent()

	headerHeight := lipgloss.Height(header)
	configHeight := lipgloss.Height(configBar)
	footerHeight := lipgloss.Height(footer)

	contentHeight := m.height - headerHeight - configHeight - footerHeight
	contentStyle := lipgloss.NewStyle().Height(contentHeight)

	fullLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		contentStyle.Render(content),
		configBar,
		footer,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, fullLayout)
}

func (m DashboardModel) renderHeader() string {
	title := titleStyle.Render("Workspace Dashboard")

	version := dimmedItemStyle.Render("v1.0.0")

	spacer := lipgloss.NewStyle().Width(m.width - lipgloss.Width(title) - lipgloss.Width(version) - 4)
	return headerStyle.Render(title + spacer.Render("") + version)
}

func (m DashboardModel) renderContent() string {
	contentHeight := m.height - 6

	leftWidth := m.width * 2 / 5
	if leftWidth < 25 {
		leftWidth = 25
	}
	if leftWidth > 50 {
		leftWidth = 50
	}
	rightWidth := m.width - leftWidth - 4

	leftStyle := panelStyle.Width(leftWidth).Height(contentHeight)
	rightStyle := panelStyle.Width(rightWidth).Height(contentHeight)

	if m.focusedPane == PaneWorkspaceList {
		leftStyle = focusedPanelStyle.Width(leftWidth).Height(contentHeight)
	} else {
		rightStyle = focusedPanelStyle.Width(rightWidth).Height(contentHeight)
	}

	leftPanel := leftStyle.Render(m.workspaceList.View())
	rightPanel := rightStyle.Render(m.details.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (m DashboardModel) renderConfigBar() string {
	if m.configManager == nil {
		return ""
	}

	cfg := m.configManager.GetConfig()

	info := dimmedItemStyle.Render(
		"Workspaces: " + shortenPath(cfg.WorkspacesDir) +
			"    Repos: " + shortenPath(cfg.ReposDir),
	)

	if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		info += "    " + statusStyle.Render("⚠ "+m.statusMessage)
	}

	return configBarStyle.Render(info)
}

func (m DashboardModel) renderFooter() string {
	sep := helpSeparatorStyle.Render("  │  ")

	keys := []string{
		helpKeyStyle.Render("↑/↓ j/k") + helpDescStyle.Render(" navigate"),
		helpKeyStyle.Render("←/→ h/l") + helpDescStyle.Render(" panels"),
		helpKeyStyle.Render("Enter") + helpDescStyle.Render(" select"),
		helpKeyStyle.Render("R") + helpDescStyle.Render(" refresh"),
		helpKeyStyle.Render("f") + helpDescStyle.Render(" fetch"),
		helpKeyStyle.Render("p") + helpDescStyle.Render(" pull"),
		helpKeyStyle.Render("d") + helpDescStyle.Render(" delete"),
		helpKeyStyle.Render("S") + helpDescStyle.Render(" settings"),
		helpKeyStyle.Render("?") + helpDescStyle.Render(" help"),
		helpKeyStyle.Render("q") + helpDescStyle.Render(" quit"),
	}

	content := ""
	for i, k := range keys {
		if i > 0 {
			content += sep
		}
		content += k
	}

	style := footerBannerStyle.Width(m.width)
	return style.Render(content)
}

// SelectedPath returns the path of the workspace the user wants to switch to.
func (m DashboardModel) SelectedPath() string {
	return m.selectedPath
}

// OpenSetup returns true if the user requested to open the setup wizard.
func (m DashboardModel) OpenSetup() bool {
	return m.openSetup
}

// DashboardResult contains the result of running the dashboard.
type DashboardResult struct {
	SelectedPath string
	OpenSetup    bool
}

// RunDashboard runs the dashboard and returns the result.
func RunDashboard(wm *workspace.Manager, cm *config.ConfigManager) (DashboardResult, error) {
	m := NewDashboard(wm, cm)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return DashboardResult{}, err
	}

	dm := finalModel.(DashboardModel)
	return DashboardResult{
		SelectedPath: dm.SelectedPath(),
		OpenSetup:    dm.OpenSetup(),
	}, nil
}
