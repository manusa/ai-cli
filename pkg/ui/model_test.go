package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/test"
	"strings"
	"testing"
	"time"
)

type testContext struct {
	m             Model
	tm            *teatest.TestModel
	SynchronizeUi bool
	llm           *test.ChatModel
}

func (c *testContext) beforeEach(t *testing.T) {
	t.Helper()
	llm := c.llm
	if llm == nil {
		llm = &test.ChatModel{}
	}
	aiAgent := ai.New(llm, config.New())
	if err := aiAgent.Run(t.Context()); err != nil {
		t.Fatalf("failed to run AI: %v", err)
	}
	c.m = NewModel(aiAgent)
	c.tm = teatest.NewTestModel(t, c.m, teatest.WithInitialTermSize(80, 20))
	teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool { return strings.Contains(string(b), "Welcome to the AI CLI!") })
	// Agent-UI synchronization
	if c.SynchronizeUi {
		go func() {
			for {
				select {
				case <-t.Context().Done():
					return
				case msg, ok := <-aiAgent.Output:
					if !ok {
						return
					}
					c.tm.Send(msg)
				}
			}
		}()
	}
}

func (c *testContext) afterEach() {
	_ = c.tm.Quit()
}

func testCase(t *testing.T, test func(c *testContext)) {
	testCaseWithContext(t, &testContext{SynchronizeUi: true}, test)
}

func testCaseWithContext(t *testing.T, ctx *testContext, test func(c *testContext)) {
	ctx.beforeEach(t)
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

func TestClear(t *testing.T) {
	testCase(t, func(c *testContext) {
		c.tm.Type("Hello AItana")
		c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "👤 ")
		})
		c.tm.Type("/clear")
		c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

		t.Run("resets viewport", func(t *testing.T) {
			// Set a term size to force viewport rendering
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Welcome to the AI CLI!")
			})
		})
		t.Run("resets composer", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "How can I help you today?")
			})
		})
	})
}

func TestViewport(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("Viewport shows welcome message", func(t *testing.T) {
			// Set a term size to force viewport rendering
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Welcome to the AI CLI!")
			})
		})
	})
	testCase(t, func(c *testContext) {
		c.tm.Send(tea.WindowSizeMsg{Width: 20, Height: 17})
		c.tm.Type("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20")
		teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "│20              │") // clear buffer
		})
		c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
		t.Run("AI notification scrolls viewport to bottom", func(t *testing.T) {
			expectedViewport := "" +
				" 18                 \r\n" +
				" 19                 \r\n" +
				" 20                 \r\n" +
				" 🤖 AI is not       \r\n" +
				" running, this is a \r\n" +
				" test"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport) &&
					!strings.Contains(string(b), " 👤 1               \r\n")
			})
		})
		t.Run("PgUp scrolls viewport one page up", func(t *testing.T) {
			c.tm.Send(tea.KeyMsg{Type: tea.KeyPgUp})

			expectedViewport := "" +
				" 👤 1               \r\n" +
				" 2                  \r\n" +
				" 3                  \r\n" +
				" 4                  \r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport) &&
					!strings.Contains(string(b), "🤖")
			})
		})
	})
}

func TestComposer(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("Composer shows placeholder text", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "How can I help you today?")
			})
		})
		t.Run("Composer has rounded borders", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
			expectedTextArea := "" +
				" ╭──────────────────────────╮ \r\n" +
				" │How can I help you today? │ \r\n" +
				" │                          │ \r\n" +
				" ╰──────────────────────────╯ \r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedTextArea)
			})
		})
		t.Run("Composer is focused and ready to receive input", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			c.tm.Type("GREETINGS PROFESSOR FALKEN")

			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "│GREETINGS PROFESSOR FALKEN")
			})
		})
		t.Run("Composer wraps text when it exceeds width", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 23, Height: 24})

			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), " │GREETINGS          │ ") &&
					strings.Contains(string(b), " │PROFESSOR FALKEN   │ ")
			})
		})
	})
}

func TestFooter(t *testing.T) {
	testCase(t, func(c *testContext) {
		// Set a term size to force footer rendering
		c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
		t.Run("Footer displays version", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.HasSuffix(string(b), "0.0.0 \u001B[80D")
			})
		})
	})
}
