package ui

import (
	"errors"
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/suite"
)

// TODO: sample PoC to build some interaction
type ModelInteractionsSuite struct {
	BaseSuite
}

func (s *ModelInteractionsSuite) TestUserMessage() {
	s.Run("User types message with enter and is sent to AI", func() {
		s.TM.Type("Hello AItana")
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Hello AItana")
		})
		s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "👤 Hello AItana")
		})
	})
}

func (s *ModelInteractionsSuite) TestErrorMessage() {
	s.Llm.StreamReader = func(_ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
		return nil, errors.New("error generating response")
	}
	s.Run("AI returns an error", func() {
		s.TM.Type("Hello Alex")
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Hello Alex")
		})
		s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		expectedViewport := "" +
			" 👤 Hello Alex\r\r\n" +
			" ❗ [NodeRunError]\r\r\n" +
			" error generating response\r\r\n"
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), expectedViewport)
		})
	})
}

func (s *ModelInteractionsSuite) TestToolMessage() {
	if runtime.GOOS == "windows" {
		s.T().Skip("Skipping test in windows") // TODO: Check, windows seems to have rendering issues
	}
	toolRequested := false
	s.Llm.StreamReader = func(_ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
		// First invocation returns a message with a tool call
		msg := schema.AssistantMessage("The list of files", []schema.ToolCall{
			{ID: "1337", Function: schema.FunctionCall{Name: "file_list"}},
		})
		// Second invocation returns the assistant's message after processing the tool call
		if toolRequested {
			msg = schema.AssistantMessage("Here is the list of files", nil)
		}
		toolRequested = true
		return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
	}
	s.Run("AI returns a tool call", func() {
		s.TM.Type("Hello Alex")
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			return strings.Contains(string(b), "Hello Alex")
		})
		s.TM.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		expectedViewport := "" +
			" 👤 Hello Alex\r\r\n" +
			"    ┌──────────────┐\r\r\n" +
			"    │ 🔧 file_list │\r\r\n" +
			"    └──────────────┘\r\r\n" +
			" 🤖 Here is the list of files"
		teatest.WaitFor(s.T(), s.TM.Output(), func(b []byte) bool {
			s.Repaint()
			return strings.Contains(string(b), expectedViewport)
		})
	})
}

func TestModelInteractions(t *testing.T) {
	suite.Run(t, new(ModelInteractionsSuite))
}
