package ai

import (
	"errors"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/suite"
)

type AiPromptSuite struct {
	suite.Suite
	Llm *test.ChatModel
	Ai  *Ai
}

func (s *AiPromptSuite) SetupTest() {
	s.Llm = &test.ChatModel{}
	s.Ai = New(
		test.NewInferenceProvider("inference-provider", test.WithInferenceAvailable(), test.WithInferenceLlm(s.Llm)),
		[]api.ToolsProvider{test.NewToolsProvider("test-tools-provider", test.WithToolsAvailable())},
	)
	if err := s.Ai.Run(config.WithConfig(s.T().Context(), config.New())); err != nil {
		s.T().Fatalf("failed to run AI: %v", err)
	}
}

func (s *AiPromptSuite) TeardownTest() {
	s.Ai.Close()
}

func (s *AiPromptSuite) WaitForRunToComplete() {
	s.Eventually(func() bool { return !s.Ai.Session().IsRunning() }, 10*time.Second, 100*time.Millisecond, "Expected AI session to finish")
}

func (s *AiPromptSuite) TestInput_SendsPrompt() {
	s.Ai.Input() <- api.NewUserMessage("Hello AItana!")

	s.WaitForRunToComplete()
	s.Run("Stores input prompt as user message in session as first message", func() {
		s.GreaterOrEqual(len(s.Ai.Session().Messages()), 1)
		s.Equal("user", s.Ai.Session().Messages()[0].Role())
		s.Equal("Hello AItana!", s.Ai.Session().Messages()[0].Text)
	})
}

func (s *AiPromptSuite) TestInput_SendsPrompt_SetUpAgentError() {
	s.Llm.WithToolsFunc = func(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
		return s.Llm, errors.New("error setting up tools")
	}
	s.Ai.Input() <- api.NewUserMessage("I will trigger an error when setting up the prompt")
	s.WaitForRunToComplete()
	s.Run("Sets error message in session if agent setup fails", func() {
		s.GreaterOrEqual(len(s.Ai.Session().Messages()), 1)
		s.Contains(s.Ai.Session().Messages(), api.NewErrorMessage("error setting up tools"))
	})
}

func (s *AiPromptSuite) TestInput_SendsPrompt_ReceivesAssistantMessage() {
	s.Llm.StreamReader = func(input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
		return schema.StreamReaderFromArray([]*schema.Message{
			schema.AssistantMessage("Hello, I am AItana!", nil),
		}), nil
	}
	s.Ai.Input() <- api.NewUserMessage("Hello AItana!")

	s.WaitForRunToComplete()
	s.Run("Stores assistant message in session ", func() {
		s.GreaterOrEqual(len(s.Ai.Session().Messages()), 2)
		s.Contains(s.Ai.Session().Messages(), api.NewAssistantMessage("Hello, I am AItana!"))
	})
}

func (s *AiPromptSuite) TestInput_SendsPrompt_ReceivesAssistantStreamedMessage() {
	s.Llm.StreamReader = func(input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
		return schema.StreamReaderFromArray([]*schema.Message{
			schema.AssistantMessage("Hello, ", nil),
			schema.AssistantMessage("I am AItana!", nil),
		}), nil
	}
	s.Ai.Input() <- api.NewUserMessage("Hello AItana!")

	s.WaitForRunToComplete()
	s.Run("Stores streamed assistant message in session (concat)", func() {
		s.GreaterOrEqual(len(s.Ai.Session().Messages()), 2)
		s.Contains(s.Ai.Session().Messages(), api.NewAssistantMessage("Hello, I am AItana!"))
	})
}

func (s *AiPromptSuite) TestInput_SendsPrompt_WithSessionMessages() {
	invocation := 0
	assistantMessages := [][]*schema.Message{
		{schema.AssistantMessage("Hello, how can I help you?", nil)},
		{schema.AssistantMessage("Yes, let me do that", nil)},
	}
	var receivedMessages []*schema.Message
	s.Llm.StreamReader = func(input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
		receivedMessages = input
		ret := schema.StreamReaderFromArray(assistantMessages[invocation])
		invocation++
		return ret, nil
	}
	s.Ai.Input() <- api.NewUserMessage("Hello AItana!")
	s.WaitForRunToComplete()
	s.Ai.Input() <- api.NewUserMessage("Help me save the world")
	s.WaitForRunToComplete()
	s.Ai.Input() <- api.NewUserMessage("Thank you!")
	s.WaitForRunToComplete()

	s.Run("Sends previous session messages as context to LLM", func() {
		s.GreaterOrEqual(len(receivedMessages), 3)
		s.Equal(schema.RoleType("user"), receivedMessages[0].Role)
		s.Equal("Hello AItana!", receivedMessages[0].Content)
		s.Equal(schema.RoleType("assistant"), receivedMessages[1].Role)
		s.Equal("Hello, how can I help you?", receivedMessages[1].Content)
		s.Equal(schema.RoleType("user"), receivedMessages[2].Role)
		s.Equal("Help me save the world", receivedMessages[2].Content)
	})
}

func TestAiPrompt(t *testing.T) {
	suite.Run(t, new(AiPromptSuite))
}
