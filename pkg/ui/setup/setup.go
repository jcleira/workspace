// Package setup provides the first-run setup wizard for the workspace CLI.
package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jcleira/workspace/pkg/config"
)

type step int

const (
	stepWelcome step = iota
	stepReposDir
	stepWorkspacesDir
	stepClaudeDir
	stepConfirm
	stepComplete
)

type SetupResult struct {
	Completed bool
	Config    *config.Config
}

type SetupModel struct {
	configManager   *config.ConfigManager
	step            step
	reposInput      DirectoryInputModel
	workspacesInput DirectoryInputModel
	claudeInput     DirectoryInputModel
	width           int
	height          int
	err             error
	quitting        bool
}

func NewSetupModel(cm *config.ConfigManager) SetupModel {
	homeDir := getHomeDir()

	return SetupModel{
		configManager:   cm,
		step:            stepWelcome,
		reposInput:      NewDirectoryInputModel("~/repos", homeDir+"/repos"),
		workspacesInput: NewDirectoryInputModel("~/workspaces", homeDir+"/workspaces"),
		claudeInput:     NewDirectoryInputModel("~/.claude", homeDir+"/.claude"),
	}
}

func NewSetupModelWithDefaults(cm *config.ConfigManager, reposDir, workspacesDir, claudeDir string) SetupModel {
	return SetupModel{
		configManager:   cm,
		step:            stepWelcome,
		reposInput:      NewDirectoryInputModel("~/repos", contractPath(reposDir)),
		workspacesInput: NewDirectoryInputModel("~/workspaces", contractPath(workspacesDir)),
		claudeInput:     NewDirectoryInputModel("~/.claude", contractPath(claudeDir)),
	}
}

func contractPath(path string) string {
	homeDir := getHomeDir()
	if len(path) > len(homeDir) && path[:len(homeDir)] == homeDir {
		return "~" + path[len(homeDir):]
	}
	return path
}

func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) && m.step != stepComplete {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.step {
		case stepWelcome:
			return m.handleWelcomeKey(msg)
		case stepReposDir:
			input, newStep, cmd := handleDirectoryKey(msg, m.reposInput, m.step, stepWelcome, stepWorkspacesDir)
			m.reposInput = input
			m.step = newStep
			if newStep == stepWorkspacesDir {
				return m, m.workspacesInput.FocusCmd()
			}
			return m, cmd
		case stepWorkspacesDir:
			input, newStep, cmd := handleDirectoryKey(msg, m.workspacesInput, m.step, stepReposDir, stepClaudeDir)
			m.workspacesInput = input
			m.step = newStep
			if newStep == stepClaudeDir {
				return m, m.claudeInput.FocusCmd()
			}
			return m, cmd
		case stepClaudeDir:
			input, newStep, cmd := handleDirectoryKey(msg, m.claudeInput, m.step, stepWorkspacesDir, stepConfirm)
			m.claudeInput = input
			m.step = newStep
			return m, cmd
		case stepConfirm:
			return m.handleConfirmKey(msg)
		case stepComplete:
			return m.handleCompleteKey(msg)
		}

	case configSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.step = stepComplete
		return m, nil

	default:
		var cmd tea.Cmd
		switch m.step { //nolint:exhaustive
		case stepReposDir:
			m.reposInput, cmd = m.reposInput.Update(msg)
		case stepWorkspacesDir:
			m.workspacesInput, cmd = m.workspacesInput.Update(msg)
		case stepClaudeDir:
			m.claudeInput, cmd = m.claudeInput.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m SetupModel) handleWelcomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Enter) {
		m.step = stepReposDir
		return m, m.reposInput.FocusCmd()
	}
	return m, nil
}

func handleDirectoryKey(msg tea.KeyMsg, input DirectoryInputModel, currentStep, prevStep, nextStep step) (DirectoryInputModel, step, tea.Cmd) {
	if msg.Type == tea.KeyTab || key.Matches(msg, keys.Tab) {
		updated, cmd := input.Update(msg)
		return updated, currentStep, cmd
	}

	if input.Mode() == pickerMode {
		if key.Matches(msg, keys.Up) || key.Matches(msg, keys.Down) ||
			key.Matches(msg, keys.Enter) || key.Matches(msg, keys.PageUp) || key.Matches(msg, keys.PageDown) {
			updated, cmd := input.Update(msg)
			return updated, currentStep, cmd
		}
	}

	switch {
	case key.Matches(msg, keys.Enter):
		if input.Mode() == textMode && input.IsValid() {
			return input, nextStep, nil
		}
		return input, currentStep, nil

	case key.Matches(msg, keys.Back):
		if input.Mode() == pickerMode {
			updated, cmd := input.Update(msg)
			return updated, currentStep, cmd
		}
		return input, prevStep, nil

	default:
		updated, cmd := input.Update(msg)
		return updated, currentStep, cmd
	}
}

func (m SetupModel) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		return m, m.saveConfig()
	case key.Matches(msg, keys.Back):
		m.step = stepClaudeDir
		return m, nil
	}
	return m, nil
}

func (m SetupModel) handleCompleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Enter) || key.Matches(msg, keys.Quit) {
		return m, tea.Quit
	}
	return m, nil
}

func (m SetupModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		reposDir := m.reposInput.Value()
		workspacesDir := m.workspacesInput.Value()
		claudeDir := m.claudeInput.Value()

		if err := os.MkdirAll(reposDir, 0o755); err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to create repos directory: %w", err)}
		}
		if err := os.MkdirAll(workspacesDir, 0o755); err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to create workspaces directory: %w", err)}
		}
		if err := os.MkdirAll(claudeDir, 0o755); err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to create claude directory: %w", err)}
		}

		if err := m.configManager.UpdateConfig(workspacesDir, reposDir, claudeDir); err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		if err := m.configManager.SetInitialized(true); err != nil {
			return configSavedMsg{err: fmt.Errorf("failed to set initialized: %w", err)}
		}

		return configSavedMsg{err: nil}
	}
}

func (m SetupModel) View() string {
	if m.quitting {
		return ""
	}

	var content string

	switch m.step {
	case stepWelcome:
		content = m.viewWelcome()
	case stepReposDir:
		content = m.viewReposDir()
	case stepWorkspacesDir:
		content = m.viewWorkspacesDir()
	case stepClaudeDir:
		content = m.viewClaudeDir()
	case stepConfirm:
		content = m.viewConfirm()
	case stepComplete:
		content = m.viewComplete()
	}

	box := wizardStyle.Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m SetupModel) viewWelcome() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Welcome to Workspace CLI"))
	b.WriteString("\n\n")
	b.WriteString("Manage multiple git repos as isolated\nworkspaces using git worktrees.\n\n")
	b.WriteString("This setup will configure:\n\n")
	b.WriteString("• Repositories directory\n")
	b.WriteString("  Where your main git repos are cloned\n\n")
	b.WriteString("• Workspaces directory\n")
	b.WriteString("  Where isolated worktree copies are created\n\n")
	b.WriteString("• Claude directory\n")
	b.WriteString("  Shared config synced across workspaces\n\n")
	b.WriteString(helpKeyStyle.Render("Enter"))
	b.WriteString(helpStyle.Render(": Start"))

	return b.String()
}

func (m SetupModel) viewReposDir() string {
	var b strings.Builder

	b.WriteString(stepIndicatorStyle.Render("Step 1 of 3"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Repository Directory"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Where are your git repositories stored?"))
	b.WriteString("\n\n")
	b.WriteString(m.reposInput.View())
	if m.reposInput.Mode() == textMode {
		b.WriteString("\n\n")
		b.WriteString(m.directoryInputHelp())
	}

	return b.String()
}

func (m SetupModel) viewWorkspacesDir() string {
	var b strings.Builder

	b.WriteString(stepIndicatorStyle.Render("Step 2 of 3"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Workspaces Directory"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Where should workspaces be created?"))
	b.WriteString("\n\n")
	b.WriteString(m.workspacesInput.View())
	if m.workspacesInput.Mode() == textMode {
		b.WriteString("\n\n")
		b.WriteString(m.directoryInputHelp())
	}

	return b.String()
}

func (m SetupModel) viewClaudeDir() string {
	var b strings.Builder

	b.WriteString(stepIndicatorStyle.Render("Step 3 of 3"))
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Claude Directory"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Where is your shared .claude directory?"))
	b.WriteString("\n\n")
	b.WriteString(m.claudeInput.View())
	if m.claudeInput.Mode() == textMode {
		b.WriteString("\n\n")
		b.WriteString(m.directoryInputHelp())
	}

	return b.String()
}

func (m SetupModel) directoryInputHelp() string {
	var b strings.Builder
	b.WriteString(helpKeyStyle.Render("Enter"))
	b.WriteString(helpDescStyle.Render(": Continue   "))
	b.WriteString(helpKeyStyle.Render("Tab"))
	b.WriteString(helpDescStyle.Render(": Browse"))
	return b.String()
}

func (m SetupModel) viewConfirm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Configuration"))
	b.WriteString("\n\n")

	b.WriteString(summaryLabelStyle.Render("Repositories: "))
	b.WriteString(summaryValueStyle.Render(m.reposInput.Value()))
	b.WriteString("\n")
	b.WriteString(summaryLabelStyle.Render("Workspaces:   "))
	b.WriteString(summaryValueStyle.Render(m.workspacesInput.Value()))
	b.WriteString("\n")
	b.WriteString(summaryLabelStyle.Render("Claude:       "))
	b.WriteString(summaryValueStyle.Render(m.claudeInput.Value()))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(validationErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(helpKeyStyle.Render("Enter"))
	b.WriteString(helpDescStyle.Render(": Save   "))
	b.WriteString(helpKeyStyle.Render("Esc"))
	b.WriteString(helpDescStyle.Render(": Back"))

	return b.String()
}

func (m SetupModel) viewComplete() string {
	var b strings.Builder

	b.WriteString(successTitleStyle.Render("Setup Complete!"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Your workspace CLI is now configured."))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Next steps:"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  1. Clone repositories to " + m.reposInput.Value()))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  2. Run 'workspace create <name>' to create a workspace"))
	b.WriteString("\n\n")
	b.WriteString(helpKeyStyle.Render("Enter"))
	b.WriteString(helpDescStyle.Render(": Continue"))

	return b.String()
}

func RunSetupWizard(cm *config.ConfigManager) (SetupResult, error) {
	model := NewSetupModel(cm)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInputTTY())

	finalModel, err := p.Run()
	if err != nil {
		return SetupResult{}, err
	}

	m := finalModel.(SetupModel)
	if m.quitting {
		return SetupResult{Completed: false}, nil
	}

	return SetupResult{
		Completed: m.step == stepComplete,
		Config:    cm.GetConfig(),
	}, nil
}

func RunSetupWizardWithDefaults(cm *config.ConfigManager, reposDir, workspacesDir, claudeDir string) (SetupResult, error) {
	model := NewSetupModelWithDefaults(cm, reposDir, workspacesDir, claudeDir)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInputTTY())

	finalModel, err := p.Run()
	if err != nil {
		return SetupResult{}, err
	}

	m := finalModel.(SetupModel)
	if m.quitting {
		return SetupResult{Completed: false}, nil
	}

	return SetupResult{
		Completed: m.step == stepComplete,
		Config:    cm.GetConfig(),
	}, nil
}
