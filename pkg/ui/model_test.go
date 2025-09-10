package ui

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/suite"
)

// AgentTuiSync synchronizes the AI agent's output with the TUI test model.
func AgentTuiSync(t *testing.T, tm *teatest.TestModel, aiAgent *ai.Ai) func() {
	return func() {
		for {
			select {
			case <-t.Context().Done():
				return
			case msg, ok := <-aiAgent.Output:
				if !ok {
					return
				}
				tm.Send(msg)
			}
		}
	}
}

type BaseSuite struct {
	suite.Suite
	Llm *test.ChatModel
	TM  *teatest.TestModel

	model *Model
}

func (s *BaseSuite) Repaint() {
	_, _ = io.ReadAll(s.TM.Output()) // Clear output buffer
	prevWidth := s.model.context.Width
	prevHeight := s.model.context.Height
	s.TM.Send(tea.WindowSizeMsg{Width: 5, Height: 5})
	s.TM.Send(tea.WindowSizeMsg{Width: prevWidth, Height: prevHeight})
}

func (s *BaseSuite) SetupTest() {
	_ = os.Setenv("TEA_STANDARD_RENDERER", "true")
	s.Llm = &test.ChatModel{}
	toolsProvider := test.NewToolsProvider("test-tools-provider", test.WithToolsAvailable())
	toolsProvider.Tools = []*api.Tool{
		{
			Name:        "file_list",
			Description: "A test tool",
			Function: func(args map[string]interface{}) (string, error) {
				return "file1.txt, file2.txt, file3.txt", nil
			},
		},
	}
	aiAgent := ai.New(&test.InferenceProvider{
		BasicInferenceProvider: api.BasicInferenceProvider{
			BasicInferenceAttributes: api.BasicInferenceAttributes{
				BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "inference-provider"},
			},
		},
		Llm: s.Llm,
	}, []api.ToolsProvider{toolsProvider})
	ctx := config.WithConfig(s.T().Context(), config.New())
	if err := aiAgent.Run(ctx); err != nil {
		s.T().Fatalf("failed to run AI: %v", err)
	}
	s.model = NewModel(aiAgent)
	s.TM = teatest.NewTestModel(s.T(), s.model, teatest.WithInitialTermSize(80, 24))
	teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool { return strings.Contains(string(b), "Welcome to the AI CLI!") })
	s.Repaint()
	go AgentTuiSync(s.T(), s.TM, aiAgent)()
}

func (s *BaseSuite) TearDownTest() {
	_ = s.TM.Quit()
	_ = os.Setenv("TEA_STANDARD_RENDERER", "")
}

type ModelSuite struct {
	BaseSuite
}

func (s *ModelSuite) TestExit() {
	s.Run("Exit with /quit", func() {
		s.TM.Type("/quit")
		s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		s.TM.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
	})
	cases := []struct{ key tea.KeyPressMsg }{
		{key: tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}},
		{key: tea.KeyPressMsg{Code: tea.KeyEsc}},
	}
	for _, tc := range cases {
		s.Run("Exit with "+tc.key.String(), func() {
			s.TM.Send(tc.key)

			s.TM.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
		})
	}
}

func (s *ModelSuite) TestClear() {
	s.TM.Type("Hello AItana")
	s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "👤 ")
	})
	s.TM.Type("/clear")
	s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

	s.Run("resets viewport", func() {
		s.Repaint()
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return !strings.Contains(string(b), "/clear") &&
				strings.Contains(string(b), "Welcome to the AI CLI!")
		})
	})
	s.Run("resets composer", func() {
		s.Repaint()
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "How can I help you today?")
		})
	})
}

func (s *ModelSuite) TestTerminalSizeWarning() {
	s.Run("shows warning when terminal height is too small", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 30, Height: 9})
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Terminal size is too small.") &&
				strings.Contains(string(b), "Minimum size is 30x10.")
		})
	})
	s.Run("shows warning when terminal width is too small", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 29, Height: 24})
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Terminal size is too small") &&
				strings.Contains(string(b), "Minimum size is 30x10.")
		})
	})
	s.Run("does not show warning when terminal size is sufficient", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 30, Height: 10})
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return !strings.Contains(string(b), "Terminal size is too small") &&
				strings.Contains(string(b), "Welcome to the AI CLI!")
		})
	})
}

func (s *ModelSuite) TestViewport() {
	s.Run("Viewport shows welcome message", func() {
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Welcome to the AI CLI!")
		})
	})
	s.TM.Send(tea.WindowSizeMsg{Width: 30, Height: 17})
	s.TM.Type("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14\n15\n16\n17\n18\n19\n20")
	teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "│20                        │") // clear buffer
	})
	s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	s.Run("AI notification scrolls viewport to bottom", func() {
		expectedViewport := "" +
			"    18\r\r\n" +
			"    19\r\r\n" +
			"    20\r\r\n" +
			" 🤖 AI is not running, this  \r\r\n" +
			"    is a test                \r\r\n"
		s.Repaint()
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), expectedViewport) &&
				!strings.Contains(string(b), "👤 1")
		})
	})
	s.Run("PgUp scrolls viewport one page up", func() {
		s.TM.Send(tea.KeyPressMsg{Code: tea.KeyPgUp})

		expectedViewport := "" +
			" 👤 1\r\r\n" +
			"    2\r\r\n" +
			"    3\r\r\n" +
			"    4\r\r\n"
		s.Repaint()
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), expectedViewport) &&
				!strings.Contains(string(b), "🤖")
		})
	})
}

func (s *ModelSuite) TestComposer() {
	s.Run("Composer shows placeholder text", func() {
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "How can I help you today?")
		})
	})
	s.Run("Composer has rounded borders", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
		expectedTextArea := "" +
			" ╭──────────────────────────╮\r\r\n" +
			" │How can I help you today? │\r\r\n" +
			" │                          │\r\r\n" +
			" ╰──────────────────────────╯\r\r\n"
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), expectedTextArea)
		})
	})
	s.TM.Type("GREETINGS PROFESSOR FALKEN")
	s.Run("Composer is focused and ready to receive input", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "│GREETINGS PROFESSOR FALKEN")
		})
	})
	s.Run("Composer wraps text when it exceeds width", func() {
		s.TM.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), " │GREETINGS PROFESSOR       │") &&
				strings.Contains(string(b), " │FALKEN                    │")
		})
	})
}

func (s *ModelSuite) TestFooter() {
	s.Run("Footer displays version", func() {
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.HasSuffix(string(b), " 0.0.0 ")
		})
	})
	s.Repaint()
	s.Run("Footer displays inference provider name", func() {
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "inference-provider")
		})
	})
}

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}
