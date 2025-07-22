package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"strings"
	"testing"
	"time"
)

type testContext struct {
	t  *testing.T
	m  Model
	tm *teatest.TestModel
}

func (c *testContext) beforeEach() {
	c.m = NewModel()
	c.tm = teatest.NewTestModel(c.t, c.m, teatest.WithInitialTermSize(80, 20))
	teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool { return strings.Contains(string(b), "Welcome to the AI CLI!") })
}

func (c *testContext) afterEach() {
	_ = c.tm.Quit()
}

func testCase(t *testing.T, test func(c *testContext)) {
	ctx := &testContext{t: t}
	ctx.beforeEach()
	t.Cleanup(ctx.afterEach)
	test(ctx)
}

func TestExit(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("Exit with /quit", func(t *testing.T) {
			c.tm.Type("/quit")
			c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

			c.tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
		})
	})
	cases := []struct{ key tea.KeyType }{
		{key: tea.KeyCtrlC},
		{key: tea.KeyEsc},
	}
	for _, tc := range cases {
		testCase(t, func(c *testContext) {
			t.Run("Exit with "+tc.key.String(), func(t *testing.T) {
				c.tm.Send(tea.KeyMsg{Type: tc.key})

				c.tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
			})
		})
	}
}

// TODO: sample PoC to build some interaction
func TestChat(t *testing.T) {
	testCase(t, func(c *testContext) {
		// Set a term size to force footer rendering
		c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
		t.Run("User types message with enter and is sent to AI", func(t *testing.T) {
			c.tm.Type("Hello AItana")
			teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Hello AItana")
			})
			c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

			teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool {
				return strings.HasPrefix(string(b), "\u001B[23A👤 Hello AItana   ")
			})
		})
	})
}

func TestComposer(t *testing.T) {
	testCase(t, func(c *testContext) {
		// Set a term size to force footer rendering
		c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
		t.Run("Composer shows placeholder text", func(t *testing.T) {
			teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "How can I help you today?")
			})
		})
		t.Run("Composer is focused and ready to receive input", func(t *testing.T) {
			c.tm.Type("GREETINGS PROFESSOR FALKEN")

			teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "GREETINGS PROFESSOR FALKEN")
			})
		})
	})
}

func TestFooter(t *testing.T) {
	testCase(t, func(c *testContext) {
		// Set a term size to force footer rendering
		c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
		t.Run("Footer displays version", func(t *testing.T) {
			teatest.WaitFor(c.t, c.tm.Output(), func(b []byte) bool {
				return strings.HasSuffix(string(b), "0.0.0 \u001B[80D")
			})
		})

	})
}
