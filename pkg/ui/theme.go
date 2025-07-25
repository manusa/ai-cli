package ui

import (
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	PrimaryBorder lipgloss.AdaptiveColor
}

var DefaultTheme = &Theme{
	PrimaryBorder: lipgloss.AdaptiveColor{Light: "013", Dark: "008"},
}

var (
	MessageToolCall = lipgloss.NewStyle().
			Margin(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(DefaultTheme.PrimaryBorder).
			Padding(0, 1)
	GlamourLightStyle = styles.LightStyleConfig
	GlamourDarkStyle  = styles.DarkStyleConfig
)

func init() {
	GlamourLightStyle.Document.Margin = uintPtr(0)
	GlamourDarkStyle.Document.Margin = uintPtr(0)
}

func uintPtr(u uint) *uint { return &u }
