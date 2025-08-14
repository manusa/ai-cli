package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/manusa/ai-cli/pkg/ui/context"
)

type Model struct {
	ctx *context.ModelContext
}

var _ tea.ViewModel = &Model{}

func NewModel(ctx *context.ModelContext) Model {
	return Model{ctx: ctx}
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
