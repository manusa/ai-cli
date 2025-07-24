package ui

import "github.com/charmbracelet/lipgloss"

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
)
