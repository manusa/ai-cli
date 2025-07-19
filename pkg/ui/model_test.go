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
	c.tm = teatest.NewTestModel(c.t, c.m)
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
		t.Run("Exit with q", func(t *testing.T) {
			c.tm.Type("q")

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
