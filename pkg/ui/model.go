package ui

import tea "github.com/charmbracelet/bubbletea"
import "github.com/charmbracelet/lipgloss"

type Model struct {
}

func NewModel() Model {
	m := Model{}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := lipgloss.NewStyle().
		Bold(true)
	return s.Render("Welcome to the AI CLI!")
}
