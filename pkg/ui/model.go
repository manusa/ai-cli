package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manusa/ai-cli/pkg/ui/components/footer"
	"github.com/manusa/ai-cli/pkg/ui/context"
	"github.com/manusa/ai-cli/pkg/version"
	"strings"
)

const composerHeight = 2

type Model struct {
	context  *context.ModelContext
	viewport viewport.Model
	composer textarea.Model
	footer   tea.Model
}

func NewModel() Model {
	ctx := &context.ModelContext{
		Version: version.Version,
	}
	m := Model{
		context:  ctx,
		viewport: viewport.New(0, 0),
		composer: textarea.New(),
		footer:   footer.NewModel(ctx),
	}
	m.composer.SetHeight(composerHeight)
	m.composer.Placeholder = "How can I help you today?"
	m.composer.FocusedStyle.CursorLine = lipgloss.NewStyle()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.viewport.Init(),
		m.footer.Init(),
		textarea.Blink,
		m.composer.Focus(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		}
	case tea.WindowSizeMsg:
		m.context.Width = msg.Width
		m.context.Height = msg.Height
		m.composer.SetWidth(msg.Width)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - composerHeight - lipgloss.Height(m.footer.View())
	}
	if m.context.Chat == "" {
		m.viewport.SetContent(lipgloss.NewStyle().Bold(true).Render("Welcome to the AI CLI!"))
	} else {
		m.viewport.SetContent(m.context.Chat)
	}
	cmds = append(cmds, m.composer.Focus())
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	m.composer, cmd = m.composer.Update(msg)
	cmds = append(cmds, cmd)
	m.footer, cmd = m.footer.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	view := strings.Builder{}
	view.WriteString(m.viewport.View() + "\n")
	view.WriteString(m.composer.View() + "\n")
	view.WriteString(m.footer.View())
	return view.String()
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	v := m.composer.Value()
	if v == "" {
		return m, nil
	}
	if v == "/quit" {
		return m, tea.Quit
	}
	m.composer.Reset()
	if m.context.Chat != "" {
		m.context.Chat += "\n"
	}
	m.context.Chat += "👤 " + v
	m.viewport.GotoBottom()
	return m, nil
}
