package setup

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}

type inputMode int

const (
	textMode inputMode = iota
	pickerMode
)

type DirectoryInputModel struct {
	textInput    textinput.Model
	mode         inputMode
	pickerPath   string
	pickerItems  []string
	pickerCursor int
	pickerScroll int
	maxVisible   int
	validation   validationState
	focused      bool
}

type validationState struct {
	valid   bool
	exists  bool
	message string
}

func NewDirectoryInputModel(placeholder, defaultValue string) DirectoryInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(defaultValue)
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	homeDir := getHomeDir()

	m := DirectoryInputModel{
		textInput:    ti,
		mode:         textMode,
		pickerPath:   homeDir,
		pickerItems:  []string{},
		pickerCursor: 0,
		pickerScroll: 0,
		maxVisible:   8,
		focused:      true,
	}

	m.updateValidation()
	m.loadPickerItems()

	return m
}

func (m *DirectoryInputModel) Focus() {
	m.focused = true
	m.textInput.Focus()
}

func (m *DirectoryInputModel) FocusCmd() tea.Cmd {
	m.focused = true
	return m.textInput.Focus()
}

func (m *DirectoryInputModel) Blur() {
	m.focused = false
	m.textInput.Blur()
}

func (m *DirectoryInputModel) Value() string {
	return m.expandPath(m.textInput.Value())
}

func (m *DirectoryInputModel) SetValue(val string) {
	m.textInput.SetValue(val)
	m.updateValidation()
}

func (m *DirectoryInputModel) IsValid() bool {
	return m.validation.valid
}

func (m DirectoryInputModel) Update(msg tea.Msg) (DirectoryInputModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.mode == textMode {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			m.updateValidation()
			return m, cmd
		}
		return m, nil
	}

	if keyMsg.Type == tea.KeyTab || key.Matches(keyMsg, keys.Tab) {
		m.toggleMode()
		return m, nil
	}

	if m.mode == pickerMode {
		return m.handlePickerKey(keyMsg)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.updateValidation()
	return m, cmd
}

func (m *DirectoryInputModel) toggleMode() {
	if m.mode == textMode {
		m.mode = pickerMode
		expanded := m.expandPath(m.textInput.Value())
		if expanded != "" {
			if info, err := os.Stat(expanded); err == nil && info.IsDir() {
				m.pickerPath = expanded
			}
		}
		m.loadPickerItems()
		m.pickerCursor = 0
		m.pickerScroll = 0
	} else {
		m.mode = textMode
		m.textInput.Focus()
		m.textInput.CursorEnd()
	}
}

func (m DirectoryInputModel) handlePickerKey(msg tea.KeyMsg) (DirectoryInputModel, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.mode = textMode
		m.textInput.Focus()
		m.textInput.CursorEnd()
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.pickerCursor > 0 {
			m.pickerCursor--
			if m.pickerCursor < m.pickerScroll {
				m.pickerScroll = m.pickerCursor
			}
		}

	case key.Matches(msg, keys.Down):
		if m.pickerCursor < len(m.pickerItems)-1 {
			m.pickerCursor++
			if m.pickerCursor >= m.pickerScroll+m.maxVisible {
				m.pickerScroll = m.pickerCursor - m.maxVisible + 1
			}
		}

	case key.Matches(msg, keys.Enter):
		if len(m.pickerItems) > 0 {
			selected := m.pickerItems[m.pickerCursor]
			if selected == ".." {
				m.pickerPath = filepath.Dir(m.pickerPath)
			} else {
				m.pickerPath = filepath.Join(m.pickerPath, selected)
			}
			m.loadPickerItems()
			m.pickerCursor = 0
			m.pickerScroll = 0
			m.textInput.SetValue(m.contractPath(m.pickerPath))
			m.updateValidation()
		}

	case key.Matches(msg, keys.PageUp):
		m.pickerCursor -= m.maxVisible
		if m.pickerCursor < 0 {
			m.pickerCursor = 0
		}
		if m.pickerCursor < m.pickerScroll {
			m.pickerScroll = m.pickerCursor
		}

	case key.Matches(msg, keys.PageDown):
		m.pickerCursor += m.maxVisible
		if m.pickerCursor >= len(m.pickerItems) {
			m.pickerCursor = len(m.pickerItems) - 1
		}
		if m.pickerCursor >= m.pickerScroll+m.maxVisible {
			m.pickerScroll = m.pickerCursor - m.maxVisible + 1
		}
	}

	return m, nil
}

func (m *DirectoryInputModel) loadPickerItems() {
	m.pickerItems = []string{}

	entries, err := os.ReadDir(m.pickerPath)
	if err != nil {
		return
	}

	homeDir := getHomeDir()
	if m.pickerPath != "/" && m.pickerPath != homeDir {
		m.pickerItems = append(m.pickerItems, "..")
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	m.pickerItems = append(m.pickerItems, dirs...)
}

func (m *DirectoryInputModel) updateValidation() {
	path := m.expandPath(m.textInput.Value())

	if path == "" {
		m.validation = validationState{
			valid:   false,
			exists:  false,
			message: "Path cannot be empty",
		}
		return
	}

	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			m.validation = validationState{
				valid:   true,
				exists:  true,
				message: "Directory exists",
			}
		} else {
			m.validation = validationState{
				valid:   false,
				exists:  true,
				message: "Path exists but is not a directory",
			}
		}
		return
	}

	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); err == nil {
		m.validation = validationState{
			valid:   true,
			exists:  false,
			message: "Directory will be created",
		}
		return
	}

	m.validation = validationState{
		valid:   true,
		exists:  false,
		message: "Directory will be created",
	}
}

func (m *DirectoryInputModel) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir := getHomeDir()
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		homeDir := getHomeDir()
		return homeDir
	}
	return path
}

func (m *DirectoryInputModel) contractPath(path string) string {
	homeDir := getHomeDir()
	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}
	return path
}

func (m DirectoryInputModel) View() string {
	var b strings.Builder

	if m.mode == textMode {
		boxStyle := inputBoxStyle
		if m.focused {
			boxStyle = focusedInputBoxStyle
		}
		b.WriteString(boxStyle.Render(m.textInput.View()))
		b.WriteString("\n")

		if m.validation.message != "" {
			var style = validationSuccessStyle
			prefix := "✓ "
			if !m.validation.valid {
				style = validationErrorStyle
				prefix = "✗ "
			} else if !m.validation.exists {
				style = validationWarningStyle
				prefix = "○ "
			}
			b.WriteString(style.Render(prefix + m.validation.message))
		}
	} else {
		b.WriteString(selectedPathStyle.Render("📁 Browse: " + m.contractPath(m.pickerPath)))
		b.WriteString("\n")

		if len(m.pickerItems) == 0 {
			b.WriteString(inputBoxStyle.Render("(no subdirectories)"))
		} else {
			visibleStart := m.pickerScroll
			visibleEnd := min(visibleStart+m.maxVisible, len(m.pickerItems))

			var pickerContent strings.Builder
			for i := visibleStart; i < visibleEnd; i++ {
				item := m.pickerItems[i]
				cursor := "  "
				style := pickerItemStyle
				if i == m.pickerCursor {
					cursor = "> "
					style = pickerSelectedStyle
				}
				if item != ".." {
					style = pickerDirStyle
				}
				pickerContent.WriteString(cursor + style.Render(item) + "\n")
			}

			b.WriteString(inputBoxStyle.Render(strings.TrimRight(pickerContent.String(), "\n")))
		}
		b.WriteString("\n")
		b.WriteString(helpDescStyle.Render("↑/↓: navigate  Enter: open  Tab: select"))
	}

	return b.String()
}

func (m DirectoryInputModel) Mode() inputMode {
	return m.mode
}
