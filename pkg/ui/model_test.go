package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/manusa/ai-cli/pkg/tools/fs"
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
	aiAgent := ai.New(llm, []*api.Tool{fs.FileList}, config.New())
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
			c.tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

			c.tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
		})
	})
	cases := []struct{ key tea.KeyPressMsg }{
		{key: tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
		{key: tea.KeyPressMsg{Code: tea.KeyEsc}},
	}
	for _, tc := range cases {
		testCase(t, func(c *testContext) {
			t.Run("Exit with "+tc.key.String(), func(t *testing.T) {
				c.tm.Send(tc.key)

				c.tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
			})
		})
	}
}

func TestClear(t *testing.T) {
	testCase(t, func(c *testContext) {
		c.tm.Type("Hello AItana")
		c.tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
		teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "👤 ")
		})
		c.tm.Type("/clear")
		c.tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		t.Run("resets viewport", func(t *testing.T) {
			// Set a term size to force viewport rendering
			c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Welcome to the AI CLI!")
			})
		})
		t.Run("resets composer", func(t *testing.T) {
			// Set a term size to force viewport rendering
			c.tm.Send(tea.WindowSizeMsg{Width: 78, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(StripAnsi(b), "How can I help you today?")
			})
		})
	})
}

func TestTerminalSizeWarning(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("shows warning when terminal height is too small", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 9})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Terminal size is too small.") &&
					strings.Contains(string(b), "Minimum size is 30x10.")
			})
		})
		t.Run("shows warning when terminal is width is too small", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 29, Height: 24})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Terminal size is too small") &&
					strings.Contains(string(b), "Minimum size is 30x10.")
			})
		})
		t.Run("does not show warning when terminal size is sufficient", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 10})
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return !strings.Contains(string(b), "Terminal size is too small") &&
					strings.Contains(string(b), "Welcome to the AI CLI!")
			})
		})
	})
}

func TestViewport(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("Viewport shows welcome message", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				// Set a term size to force viewport rendering
				c.tm.Send(tea.WindowSizeMsg{Width: 78, Height: 24})
				return strings.Contains(string(b), "Welcome to the AI CLI!")
			})
		})
	})
	testCase(t, func(c *testContext) {
		c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 17})
		c.tm.Type("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20")
		teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "│20                        │") // clear buffer
		})
		c.tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
		t.Run("AI notification scrolls viewport to bottom", func(t *testing.T) {
			expectedViewport := "" +
				"    18                        \r\n" +
				"    19                        \r\n" +
				"    20                        \r\n" +
				" 🤖 \u001B[38;5;252mAI is not running, this\u001B[38;5;252m \u001B[0m\u001B[38;5;252m \u001B[0m \r\n" +
				"    \u001B[0m\u001B[38;5;252mis a test                \u001B[0m \r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport) &&
					!strings.Contains(string(b), " 👤 1               \r\n")
			})
		})
		t.Run("PgUp scrolls viewport one page up", func(t *testing.T) {
			c.tm.Send(tea.KeyPressMsg{Code: tea.KeyPgUp})

			expectedViewport := "" +
				" 👤 1                         \r\n" +
				"    2                         \r\n" +
				"    3                         \r\n" +
				"    4                         \r\n"
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
				return strings.Contains(StripAnsi(b), "How can I help you today?")
			})
		})
		t.Run("Composer has rounded borders", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
			expectedTextArea := "" +
				" ╭──────────────────────────╮\n" +
				" │How can I help you today? │\n" +
				" │                          │\n" +
				" ╰──────────────────────────╯\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(StripAnsi(b), expectedTextArea)
			})
		})
		c.tm.Type("GREETINGS PROFESSOR FALKEN")
		t.Run("Composer is focused and ready to receive input", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
				return strings.Contains(StripAnsi(b), "│GREETINGS PROFESSOR FALKEN")
			})
		})
		t.Run("Composer wraps text when it exceeds width", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
				return strings.Contains(StripAnsi(b), " │GREETINGS PROFESSOR       │") &&
					strings.Contains(StripAnsi(b), " │FALKEN                    │")
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
				return strings.HasSuffix(string(b), "\u001B[30m0.0.0\u001B[39m \u001B[m")
			})
		})
	})
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func StripAnsi(b []byte) string {
	return re.ReplaceAllString(string(b), "")
}
