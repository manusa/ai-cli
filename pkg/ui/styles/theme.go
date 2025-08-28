package styles

import (
	"image/color"

	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/glamour/v2/ansi"
	"github.com/charmbracelet/glamour/v2/styles"
	"github.com/charmbracelet/lipgloss/v2"
)

type Theme struct {
	IsDark          bool
	PrimaryBorder   color.Color
	ComposerStyles  textarea.Styles
	GlamourStyle    ansi.StyleConfig
	MessageToolCall lipgloss.Style
}

func DefaultTheme(isDark bool) *Theme {
	lightDark := lipgloss.LightDark(isDark)
	theme := &Theme{
		IsDark:         isDark,
		PrimaryBorder:  lightDark(lipgloss.Color("013"), lipgloss.Color("008")),
		ComposerStyles: textarea.DefaultStyles(isDark),
	}
	// Composer styles
	theme.ComposerStyles.Cursor.Blink = false
	theme.ComposerStyles.Focused.CursorLine = lipgloss.NewStyle()       // Removes highlighted line
	theme.ComposerStyles.Focused.CursorLineNumber = lipgloss.NewStyle() // Removes highlighted line
	theme.ComposerStyles.Focused.Base = theme.ComposerStyles.Focused.Base.Border(lipgloss.RoundedBorder())
	// Glamour styles
	if isDark {
		theme.GlamourStyle = styles.DarkStyleConfig
	} else {
		theme.GlamourStyle = styles.LightStyleConfig
	}
	theme.GlamourStyle.Document.Margin = uintPtr(0)
	// Widget styles
	theme.MessageToolCall = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.PrimaryBorder).
		Padding(0, 1)
	return theme
}

func uintPtr(u uint) *uint { return &u }
