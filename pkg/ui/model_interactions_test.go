package ui

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/test"
	"runtime"
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

			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "👤 Hello AItana")
			})
		})
	})
}

func TestInteractionsError(t *testing.T) {
	ctx := &testContext{
		SynchronizeUi: true,
		llm: &test.ChatModel{
			StreamReader: func(_ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
				return nil, errors.New("error generating response")
			},
		},
	}
	testCaseWithContext(t, ctx, func(c *testContext) {
		c.tm.Send(tea.WindowSizeMsg{Width: 30, Height: 24})
		t.Run("AI returns an error", func(t *testing.T) {
			c.tm.Type("Hello Alex")
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Hello Alex")
			})
			c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

			expectedViewport := "" +
				" 👤 Hello Alex                \r\n" +
				" ❗ [NodeRunError]            \r\n" +
				" error generating response    \r\n"
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport)
			})
		})
	})
}

func TestInteractionsTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test in windows") // TODO: Check, windows seems to have rendering issues
	}
	toolRequested := false
	ctx := &testContext{
		SynchronizeUi: true,
		llm: &test.ChatModel{
			StreamReader: func(_ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
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
			},
		},
	}
	testCaseWithContext(t, ctx, func(c *testContext) {
		c.tm.Send(tea.WindowSizeMsg{Width: 32, Height: 30})
		t.Run("AI returns a tool call", func(t *testing.T) {
			c.tm.Type("Hello Alex")
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), "Hello Alex")
			})
			c.tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

			expectedViewport := "" +
				" 👤 Hello Alex                  \r\n" +
				"    ┌──────────────┐            \r\n" +
				"    │ 🔧 file_list │            \r\n" +
				"    └──────────────┘            \r\n" +
				" 🤖 \u001B[38;5;252mHere is the list of files  \u001B[0m "
			teatest.WaitFor(t, c.tm.Output(), func(b []byte) bool {
				return strings.Contains(string(b), expectedViewport)
			})
		})
	})
}
