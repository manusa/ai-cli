package footer

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/manusa/ai-cli/pkg/ui/context"
)

type Footer interface {
	tea.ViewModel
}

type model struct {
	ctx *context.ModelContext
}

var _ Footer = (*model)(nil)

func New(ctx *context.ModelContext) Footer {
	return model{ctx: ctx}
}

func (m model) View() string {
	style := lipgloss.NewStyle().
		Background(m.ctx.Theme.FooterBackground).
		Foreground(m.ctx.Theme.FooterText).
		Padding(0, 1)
	left := style.Render("🧠", m.ctx.Ai.InferenceAttributes().Name(),
		"🛠️", strconv.Itoa(m.ctx.Ai.ToolCount()))
	version := style.Render(m.ctx.Version)
	spacerWidth := m.ctx.Width - lipgloss.Width(left) - lipgloss.Width(version)
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := lipgloss.NewStyle().
		Background(m.ctx.Theme.FooterBackground).
		Render(strings.Repeat(" ", spacerWidth))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, version)
}
