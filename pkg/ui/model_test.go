package ui

import (
	"os"
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
	"github.com/stretchr/testify/assert"
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
	inferenceProvider := &test.InferenceProvider{
		BasicInferenceProvider: api.BasicInferenceProvider{
			BasicInferenceAttributes: api.BasicInferenceAttributes{
				BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "inference-provider"},
			},
		},
		Available: true,
		Llm:       llm,
	}
	aiAgent := ai.New(config.New(), inferenceProvider, []*api.Tool{fs.FileList})
	if err := aiAgent.Run(t.Context()); err != nil {
		t.Fatalf("failed to run AI: %v", err)
	}
	// Use standard renderer
	_ = os.Setenv("TEA_STANDARD_RENDERER", "true")
	c.m = NewModel(aiAgent)
	c.tm = teatest.NewTestModel(t, c.m, teatest.WithInitialTermSize(40, 40))
	teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool { return strings.Contains(string(b), "Welcome to the AI CLI!") })
	c.tm = teatest.NewTestModel(t, c.m, teatest.WithInitialTermSize(80, 20)) // Force repaint
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

		var buffer []byte
		teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
			buffer = b
			return !strings.Contains(string(buffer), "/clear") &&
				strings.Contains(string(buffer), "Welcome to the AI CLI!")
		})
		t.Run("resets viewport", func(t *testing.T) {
			assert.Contains(t, string(buffer), "Welcome to the AI CLI!")
		})
		t.Run("resets composer", func(t *testing.T) {
			assert.Contains(t, string(buffer), "How can I help you today?")
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
		t.Run("shows warning when terminal width is too small", func(t *testing.T) {
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
				"    18\r\r\n" +
				"    19\r\r\n" +
				"    20\r\r\n" +
				" 🤖 AI is not running, this  \r\r\n" +
				"    is a test                \r\r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport) &&
					!strings.Contains(string(b), "👤 1")
			})
		})
		t.Run("PgUp scrolls viewport one page up", func(t *testing.T) {
			c.tm.Send(tea.KeyPressMsg{Code: tea.KeyPgUp})

			expectedViewport := "" +
				" 👤 1\r\r\n" +
				"    2\r\r\n" +
				"    3\r\r\n" +
				"    4\r\r\n"
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
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "How can I help you today?")
			})
		})
		t.Run("Composer has rounded borders", func(t *testing.T) {
			c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
			expectedTextArea := "" +
				" ╭──────────────────────────╮\r\r\n" +
				" │How can I help you today? │\r\r\n" +
				" │                          │\r\r\n" +
				" ╰──────────────────────────╯\r\r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedTextArea)
			})
		})
		c.tm.Type("GREETINGS PROFESSOR FALKEN")
		t.Run("Composer is focused and ready to receive input", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
				return strings.Contains(string(b), "│GREETINGS PROFESSOR FALKEN")
			})
		})
		t.Run("Composer wraps text when it exceeds width", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
				return strings.Contains(string(b), " │GREETINGS PROFESSOR       │") &&
					strings.Contains(string(b), " │FALKEN                    │")
			})
		})
	})
}

func TestFooter(t *testing.T) {
	testCase(t, func(c *testContext) {
		t.Run("Footer displays version", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.HasSuffix(string(b), " 0.0.0 ")
			})
		})
		c.tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24}) // Force repaint
		t.Run("Footer displays inference provider name", func(t *testing.T) {
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "inference-provider")
			})
		})
	})
}
