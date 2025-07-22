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
	m.composer.ShowLineNumbers = false
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
	case ai.Notification:
		// AI is running and a new partial message is available
		// Partial message rendering is handled by the ai.Session itself
		m.viewport.GotoBottom()

	}
	// Update viewport
	if !m.context.Ai.Session().HasMessages() && !m.context.Ai.Session().IsRunning() {
		m.viewport.SetContent(lipgloss.NewStyle().Bold(true).Render("Welcome to the AI CLI!"))
	} else {
		m.viewport.SetContent(m.renderMessages())
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

func (m Model) renderMessages() string {
	renderedMessages := strings.Builder{}
	for idx, msg := range m.context.Ai.Session().Messages() {
		if idx > 0 {
			renderedMessages.WriteString("\n")
		}
		var text string
		switch msg.Type {
		case ai.MessageTypeSystem:
			text = "🤖 "
		case ai.MessageTypeUser:
			text = "👤 "
		case ai.MessageTypeAssistant:
			text = "🤖 "
		case ai.MessageTypeTool:
			text = "🔧 "
		}
		text += msg.Text
		maxWidth := m.context.Width
		// create chunks of text that fit within the maxWidth
		if len(text) > maxWidth {
			chunks := make([]string, 0)
			for len(text) > maxWidth {
				chunk := text[:maxWidth]
				chunks = append(chunks, chunk)
				text = text[maxWidth:]
			}
			if len(text) > 0 {
				chunks = append(chunks, text)
			}
			text = strings.Join(chunks, "\n")
		}
		renderedMessages.WriteString(strings.Trim(text, "\n") + "\n")
	}
	return renderedMessages.String()
	// TODO: Glamour doesn't work well
	//const glamourGutter = 2
	//glamourRenderWidth := m.context.Width - glamourGutter
	//renderer, err := glamour.NewTermRenderer(
	//	glamour.WithAutoStyle(),
	//	glamour.WithWordWrap(glamourRenderWidth),
	//)
	//if err != nil {
	//	return renderedMessages.String() // Return raw text if rendering fails
	//}
	//defer func() { _ = renderer.Close() }()
	//str, err := renderer.Render(renderedMessages.String())
	//if err != nil {
	//	return renderedMessages.String() // Return raw text if rendering fails
	//}
	//return str
}
