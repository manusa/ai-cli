package footer

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manusa/ai-cli/pkg/ui/context"
	"strings"
)

type Model struct {
	ctx *context.ModelContext
}

func NewModel(ctx *context.ModelContext) Model {
	return Model{ctx: ctx}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	version := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1).
		Render(m.ctx.Version)
	spacerWidth := m.ctx.Width - lipgloss.Width(version)
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Render(strings.Repeat(" ", spacerWidth))
	return lipgloss.JoinHorizontal(lipgloss.Top, spacer, version)
}
