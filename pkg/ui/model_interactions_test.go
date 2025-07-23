package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/manusa/ai-cli/pkg/ai"
	"strings"
	"testing"
)

// TODO: sample PoC to build some interaction
func TestInteractionsUser(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("User types message with enter and is sent to AI", func(t *testing.T) {
			c.tm.Type("Hello AItana")
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Hello AItana")
			})
			c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
			c.tm.Send(ai.Notification{}) // TODO, enable AI sync in context

			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "👤 Hello AItana")
			})
		})
	})
}
