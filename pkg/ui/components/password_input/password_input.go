package password_input

import (
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type model struct {
	textInput textinput.Model
	quitting  bool
}

func initialModel() model {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)

	return model{textInput: ti}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() (string, *tea.Cursor) {
	var c *tea.Cursor
	if !m.textInput.VirtualCursor() {
		c = m.textInput.Cursor()
	}

	str := lipgloss.JoinVertical(lipgloss.Top, m.textInput.View(), m.footerView())
	if m.quitting {
		str += "\n"
	}

	return str, c
}

func (m model) footerView() string { return "\n(esc to quit)" }

func Prompt() (string, error) {
	promptModel := initialModel()
	p := tea.NewProgram(promptModel)
	if res, err := p.Run(); err != nil {
		return "", err
	} else {
		return res.(model).textInput.Value(), nil
	}
}
