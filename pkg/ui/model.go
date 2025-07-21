package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manusa/ai-cli/pkg/ui/components/footer"
	"github.com/manusa/ai-cli/pkg/ui/context"
	"github.com/manusa/ai-cli/pkg/version"
	"strings"
)

type Model struct {
	context  *context.ModelContext
	viewport viewport.Model
	footer   tea.Model
}

func NewModel() Model {
	ctx := &context.ModelContext{
		Version: version.Version,
	}
	m := Model{
		context:  ctx,
		viewport: viewport.New(0, 0),
		footer:   footer.NewModel(ctx),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	tea.Batch(m.viewport.Init(), m.footer.Init())
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.context.Width = msg.Width
		m.context.Height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - lipgloss.Height(m.footer.View())
	}
	m.viewport.SetContent(lipgloss.NewStyle().Bold(true).Render("Welcome to the AI CLI!"))
	m.viewport, _ = m.viewport.Update(msg)
	m.footer, _ = m.footer.Update(msg)
	return m, nil

	// TODO: Probably need to check what's active and then return the cmd for that component.
	//var viewportCmd, footerCmd tea.Cmd
	//if m.footer, footerCmd = m.footer.Update(msg); footerCmd != nil {
	//	return m, footerCmd
	//}
	//if m.viewport, viewportCmd = m.viewport.Update(msg); viewportCmd != nil {
	//	return m, viewportCmd
	//}
	//return m, tea.Batch(footerCmd)
}

func (m Model) View() string {
	view := strings.Builder{}
	view.WriteString(m.viewport.View() + "\n")
	view.WriteString(m.footer.View())
	return view.String()
}
