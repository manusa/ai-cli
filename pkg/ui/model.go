package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/ui/components/footer"
	"github.com/manusa/ai-cli/pkg/ui/context"
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/muesli/termenv"
	"strings"
)

const (
	minWidth                  = 30
	minHeight                 = 10
	composerPaddingHorizontal = 1
)

type Model struct {
	context  *context.ModelContext
	viewport viewport.Model
	spinner  spinner.Model
	composer textarea.Model
	footer   tea.Model
}

func NewModel(ai *ai.Ai) Model {
	ctx := &context.ModelContext{
		Ai:                ai,
		Version:           version.Version,
		HasDarkBackground: termenv.HasDarkBackground(),
	}
	m := Model{
		context:  ctx,
		viewport: viewport.New(0, 0),
		spinner:  spinner.New(spinner.WithSpinner(spinner.Points)),
		composer: textarea.New(),
		footer:   footer.NewModel(ctx),
	}
	m.viewport.KeyMap = ViewportKeyMap()
	m.composer.SetHeight(2)
	m.composer.ShowLineNumbers = false
	m.composer.Placeholder = "How can I help you today?"
	m.composer.Prompt = ""
	m.composer.Cursor.SetMode(cursor.CursorStatic)
	m.composer.FocusedStyle.CursorLine = lipgloss.NewStyle() // Removes highlighted line
	m.composer.FocusedStyle.Base = m.composer.FocusedStyle.Base.Border(lipgloss.RoundedBorder())
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.viewport.Init(),
		m.spinner.Tick,
		m.composer.Focus(),
		m.footer.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	session := m.context.Ai.Session()
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
		m.composer.SetWidth(msg.Width - composerPaddingHorizontal*2)
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case cursor.BlinkMsg:
		// Only affects the composer textarea, minimize the impact on performance
		m.composer, cmd = m.composer.Update(msg)
		return m, cmd
	case ai.Notification:
		// AI is running and a new partial message is available
		// Partial message rendering is handled by the ai.Session itself
		m.viewport.GotoBottom()
	}
	// Update viewport
	adjustViewportSize(&m)
	if !session.HasMessages() {
		m.viewport.SetContent(lipgloss.NewStyle().Bold(true).Render("Welcome to the AI CLI!"))
	} else {
		m.viewport.SetContent(m.renderMessages())
	}

	cmds = append(cmds, m.composer.Focus())
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	if !session.IsRunning() {
		// Ignore input while AI is running
		m.composer, cmd = m.composer.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.footer, cmd = m.footer.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	center := lipgloss.NewStyle().Width(m.context.Width).AlignHorizontal(lipgloss.Center)
	if m.context.Width < minWidth || m.context.Height < minHeight {
		return center.Height(m.context.Height).AlignVertical(lipgloss.Center).
			Render("Terminal size is too small.\n" +
				"Minimum size is " + fmt.Sprintf("%dx%d.", minWidth, minHeight))
	}
	view := strings.Builder{}
	view.WriteString(m.viewport.View() + "\n")
	if m.context.Ai.Session().IsRunning() {
		view.WriteString(center.Render(m.spinner.View()) + "\n")
	}
	view.WriteString(center.Render(m.composer.View()) + "\n")
	view.WriteString(m.footer.View())
	return view.String()
}

func (m Model) handleEnter() (Model, tea.Cmd) {
	if m.context.Ai.Session().IsRunning() {
		// AI is running, ignore the input
		return m, nil
	}
	v := m.composer.Value()
	switch v {
	case "":
		return m, nil // Ignore empty input
	case "/clear":
		m.context.Ai.Reset()
		m.composer.Reset()
		m.viewport.GotoTop()
		return m, tea.ClearScreen
	case "/quit":
		return m, tea.Quit
	}
	m.composer.Reset()
	m.context.Ai.Input <- api.NewUserMessage(v)
	m.viewport.GotoBottom()
	return m, nil
}

func (m Model) renderMessages() string {
	renderedMessages := strings.Builder{}
	for idx, msg := range m.context.Ai.Session().Messages() {
		if idx > 0 {
			renderedMessages.WriteString("\n")
		}
		renderedMessages.WriteString(render(m.context, msg))
	}
	return renderedMessages.String()
}

func adjustViewportSize(m *Model) {
	spinnerHeight := 0
	if m.context.Ai.Session().IsRunning() {
		spinnerHeight = lipgloss.Height(m.spinner.View())
	}
	composerHeight := m.composer.Height() + m.composer.FocusedStyle.Base.GetVerticalFrameSize()
	m.viewport.Width = m.context.Width
	m.viewport.Height = m.context.Height - spinnerHeight - composerHeight - lipgloss.Height(m.footer.View())
}

func emoji(messageType api.MessageType) string {
	switch messageType {
	case api.MessageTypeSystem:
		return "🤖"
	case api.MessageTypeUser:
		return "👤"
	case api.MessageTypeAssistant:
		return "🤖"
	case api.MessageTypeTool:
		return "🔧"
	case api.MessageTypeError:
		return "❗"
	}
	return ">"
}

func render(context *context.ModelContext, msg api.Message) string {
	// TODO use constants for gutters
	maxWidth := context.Width
	switch msg.Type {
	case api.MessageTypeTool:
		return MessageToolCall.MaxWidth(maxWidth - 2).Render("🔧 " + msg.Text)
	case api.MessageTypeAssistant:
		glamourStyle := GlamourLightStyle
		if context.HasDarkBackground {
			glamourStyle = GlamourDarkStyle
		}
		tr, err := glamour.NewTermRenderer(
			glamour.WithStyles(glamourStyle),
			glamour.WithWordWrap(maxWidth-5),
			glamour.WithEmoji(),
		)
		defer func() { _ = tr.Close() }()
		if err != nil {
			break
		}
		if out, err := tr.Render(strings.Trim(msg.Text, "\n")); err == nil {
			out = lipgloss.NewStyle().Foreground(lipgloss.Color("#4e9a06")).Render("🤖 AI") + "\n" +
				lipgloss.NewStyle().MarginLeft(3).Render(strings.Trim(out, "\n"))
			return lipgloss.NewStyle().Width(maxWidth-2).Margin(0, 1).Render(out)
		}
	}
	messageStyle := lipgloss.NewStyle().Width(maxWidth-2).Margin(0, 1)
	return messageStyle.Render(emoji(msg.Type), strings.Trim(msg.Text, "\n"))
}
