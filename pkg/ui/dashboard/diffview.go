package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DiffViewModel struct {
	content     string
	viewport    viewport.Model
	repoName    string
	staged      bool
	loading     bool
	err         error
	width       int
	height      int
	keys        KeyMap
	ready       bool
	waitingForG bool
}

func NewDiffView(keys KeyMap) DiffViewModel {
	return DiffViewModel{
		keys: keys,
	}
}

func (m DiffViewModel) SetContent(content, repoName string, staged bool) DiffViewModel {
	m.content = content
	m.repoName = repoName
	m.staged = staged
	m.loading = false
	m.err = nil

	if m.ready {
		m.viewport.SetContent(content)
		m.viewport.GotoTop()
	}

	return m
}

func (m DiffViewModel) SetSize(width, height int) DiffViewModel {
	m.width = width
	m.height = height

	headerHeight := 3
	footerHeight := 2
	viewportHeight := height - headerHeight - footerHeight

	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.viewport = viewport.New(width-4, viewportHeight)
	m.viewport.Style = lipgloss.NewStyle()
	m.ready = true

	if m.content != "" {
		m.viewport.SetContent(m.content)
	}

	return m
}

func (m DiffViewModel) SetLoading(loading bool) DiffViewModel {
	m.loading = loading
	return m
}

func (m DiffViewModel) SetError(err error) DiffViewModel {
	m.err = err
	m.loading = false
	return m
}

func (m DiffViewModel) Init() tea.Cmd {
	return nil
}

func (m DiffViewModel) Update(msg tea.Msg) (DiffViewModel, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "k", "up":
			m.viewport.ScrollUp(1)
			m.waitingForG = false
		case "j", "down":
			m.viewport.ScrollDown(1)
			m.waitingForG = false
		case "ctrl+u":
			m.viewport.HalfPageUp()
			m.waitingForG = false
		case "ctrl+d":
			m.viewport.HalfPageDown()
			m.waitingForG = false
		case "g":
			if m.waitingForG {
				m.viewport.GotoTop()
				m.waitingForG = false
			} else {
				m.waitingForG = true
			}
		case "G":
			m.viewport.GotoBottom()
			m.waitingForG = false
		default:
			m.waitingForG = false
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DiffViewModel) View() string {
	if m.loading {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			dimmedItemStyle.Render("Loading diff..."),
		)
	}

	if m.err != nil {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			statusErrorStyle.Render("Error: "+m.err.Error()+"\n\nPress q or Esc to close"),
		)
	}

	if strings.TrimSpace(m.content) == "" {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			dimmedItemStyle.Render("No changes to display\n\nPress q or Esc to close"),
		)
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

func (m DiffViewModel) renderHeader() string {
	title := "DIFF"
	if m.staged {
		title = "STAGED DIFF"
	}

	repoInfo := fmt.Sprintf("Repository: %s", m.repoName)

	diffTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor)

	headerBox := lipgloss.NewStyle().
		Border(lipgloss.Border{Bottom: "─"}).
		BorderForeground(subtleColor).
		Width(m.width-2).
		Padding(0, 1)

	content := diffTitleStyle.Render(title) + "  " + dimmedItemStyle.Render(repoInfo)
	return headerBox.Render(content)
}

func (m DiffViewModel) renderFooter() string {
	scrollPercent := m.viewport.ScrollPercent() * 100
	position := fmt.Sprintf("%.0f%%", scrollPercent)

	help := helpKeyStyle.Render("j/k") + helpDescStyle.Render(" scroll  ") +
		helpKeyStyle.Render("gg/G") + helpDescStyle.Render(" top/bottom  ") +
		helpKeyStyle.Render("Ctrl+U/D") + helpDescStyle.Render(" page  ") +
		helpKeyStyle.Render("q/Esc") + helpDescStyle.Render(" close")

	spacerWidth := m.width - lipgloss.Width(help) - lipgloss.Width(position) - 6
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	return footerBannerStyle.Width(m.width).Render(help + spacer + position)
}
