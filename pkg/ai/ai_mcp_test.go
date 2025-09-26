package ai

import (
	"slices"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/suite"
)

type AiMcpSuite struct {
	suite.Suite
	Llm *test.ChatModel
	Ai  *Ai
}

func (s *AiMcpSuite) SetupTest() {
	s.Llm = &test.ChatModel{}
	s.Ai = New(
		test.NewInferenceProvider("inference-provider", test.WithInferenceAvailable(), test.WithInferenceLlm(s.Llm)),
		[]api.ToolsProvider{test.NewToolsProvider("test-toolManager-provider", test.WithToolsAvailable(), test.WithToolsMcpSettings(test.McpServer()))},
	)
	if err := s.Ai.Run(config.WithConfig(s.T().Context(), config.New())); err != nil {
		s.T().Fatalf("failed to run AI: %v", err)
	}
}

func (s *AiMcpSuite) TearDownTest() {
	s.Ai.Close()
}

func (s *AiMcpSuite) TestRunStartsMcpServers() {
	s.Require().NotEmpty(s.Ai.mcpClients, "Expected MCP clients to be initialized")
	s.Run("Adds MCP clients for each MCP-enabled toolManager provider", func() {
		s.Require().Len(s.Ai.mcpClients, 1, "Expected MCP clients to be initialized")
	})
	s.Run("MCP clients are initialized", func() {
		s.False(slices.ContainsFunc(s.Ai.mcpClients, func(c *ToolsProviderMcpClient) bool { return c.InitializeResult() == nil }), "Expected all MCP clients to be initialized")
	})
}

func (s *AiMcpSuite) TestCloseShutsDownMcpServers() {
	s.Require().NotEmpty(s.Ai.mcpClients, "Expected MCP clients to be initialized")
	s.Ai.Close()
	s.Run("MCP clients are shut down", func() {
		err := s.Ai.mcpClients[0].Ping(s.T().Context(), nil)
		s.Error(err, "Expected error when pinging closed MCP client")
		s.Contains(err.Error(), "closed", "Expected connection is closed error")
	})
}

func (s *AiMcpSuite) TestNewReActAgentAddsOnlyEnableTool() {
	var receivedTools []*schema.ToolInfo
	s.Llm.WithToolsFunc = func(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
		receivedTools = tools
		return s.Llm, nil
	}
	s.Ai.Input() <- api.NewUserMessage("Hello AItana! I'm sending some MCP toolManager.")
	s.Require().Eventually(func() bool { return len(receivedTools) > 0 }, 10*time.Second, 100, "Expected LLM to be called with toolManager")
	s.Run("Tools includes a single tool", func() {
		s.Len(receivedTools, 1, "Expected only one tool to be passed to LLM")
	})
	s.Run("Tools includes toolset_enable", func() {
		s.True(slices.ContainsFunc(receivedTools, func(t *schema.ToolInfo) bool { return t.Name == "toolset_enable" }), "Expected to find MCP 'toolset_enable' tool in received toolManager")
	})
}

func TestAiMcp(t *testing.T) {
	suite.Run(t, new(AiMcpSuite))
}
