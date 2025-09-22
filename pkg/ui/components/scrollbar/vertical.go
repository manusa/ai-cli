package scrollbar

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/manusa/ai-cli/pkg/ui/context"
)

type Vertical interface {
	tea.ViewModel
	Width() int
	SetHeight(height, visibleLines, totalLines, offset int)
}

type vertical struct {
	ctx         *context.ModelContext
	style       lipgloss.Style
	arrowStyle  lipgloss.Style
	thumbStyle  lipgloss.Style
	trackStyle  lipgloss.Style
	height      int
	thumbHeight int
	thumbOffset int
}

var _ Vertical = (*vertical)(nil)

// NewVertical create a new vertical scrollbar.
func NewVertical(ctx *context.ModelContext) Vertical {
	return &vertical{
		ctx:        ctx,
		style:      lipgloss.NewStyle().Width(1),
		arrowStyle: lipgloss.NewStyle().Foreground(ctx.Theme.ScrollbarForeground).Background(ctx.Theme.ScrollbarBackground),
		thumbStyle: lipgloss.NewStyle().Background(ctx.Theme.ScrollbarForeground).SetString("\u2002"),
		trackStyle: lipgloss.NewStyle().Background(ctx.Theme.ScrollbarBackground).SetString("\u2002"),
	}
}

func (m *vertical) Width() int {
	return m.style.GetWidth()
}

func (m *vertical) SetHeight(height, visibleLines, totalLines, offset int) {
	arrowsHeight := 2 // One arrow at the top and one at the bottom
	m.height = height - m.style.GetVerticalFrameSize() - arrowsHeight
	ratio := float64(visibleLines) / float64(totalLines)
	m.thumbHeight = max(1, int(math.Round(float64(m.height)*ratio)))
	m.thumbOffset = max(0, min(m.height-m.thumbHeight, int(math.Round(float64(offset)*ratio))))
}

func (m *vertical) View() string {
	if m.thumbHeight == m.height {
		return m.style.Render(strings.Repeat("\u2002\n", m.height+2))
	}
	bar := strings.TrimRight(
		m.arrowStyle.Render("▲")+
			strings.Repeat(m.trackStyle.String()+"\n", m.thumbOffset)+
			strings.Repeat(m.thumbStyle.String()+"\n", m.thumbHeight)+
			strings.Repeat(m.trackStyle.String()+"\n", max(0, m.height-m.thumbOffset-m.thumbHeight))+
			m.arrowStyle.Render("▼"),
		"\n",
	)
	return m.style.Render(bar)
}
