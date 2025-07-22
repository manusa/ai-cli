package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manusa/ai-cli/pkg/ai"
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

func NewModel(ai *ai.Ai) Model {
	ctx := &context.ModelContext{
		Ai:      ai,
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
	if len(m.context.Ai.Session().Messages) == 0 {
		m.viewport.SetContent(lipgloss.NewStyle().Bold(true).Render("Welcome to the AI CLI!"))
	} else {
		m.viewport.SetContent(render(m.context.Ai.Session().Messages))
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
	m.context.Ai.Input <- ai.NewUserMessage(v)
	m.viewport.GotoBottom()
	return m, nil
}

func render(messages []ai.Message) string {
	renderedMessages := strings.Builder{}
	for _, msg := range messages {
		switch msg.Type {
		case ai.MessageTypeSystem:
			renderedMessages.WriteString("🤖 " + msg.Text + "\n")
		case ai.MessageTypeUser:
			renderedMessages.WriteString("👤 " + msg.Text + "\n")
		case ai.MessageTypeAssistant:
			renderedMessages.WriteString("🤖 " + msg.Text + "\n")
		case ai.MessageTypeTool:
			renderedMessages.WriteString("🔧 " + msg.Text + "\n")
		default:
			renderedMessages.WriteString(msg.Text + "\n")
		}
	}
	return renderedMessages.String()
}
