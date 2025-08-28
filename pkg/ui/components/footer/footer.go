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
	style := lipgloss.NewStyle().
		Background(m.ctx.Theme.FooterBackground).
		Foreground(m.ctx.Theme.FooterText).
		Padding(0, 1)
	inferenceProvider := style.Render("🧠", m.ctx.Ai.InferenceAttributes().Name())
	version := style.Render(m.ctx.Version)
	spacerWidth := m.ctx.Width - lipgloss.Width(inferenceProvider) - lipgloss.Width(version)
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := lipgloss.NewStyle().
		Background(m.ctx.Theme.FooterBackground).
		Render(strings.Repeat(" ", spacerWidth))
	return lipgloss.JoinHorizontal(lipgloss.Top, inferenceProvider, spacer, version)
}
