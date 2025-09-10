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
	"github.com/mark3labs/mcp-go/client"
	m3lmcp "github.com/mark3labs/mcp-go/mcp"
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
		[]api.ToolsProvider{test.NewToolsProvider("test-tools-provider", test.WithToolsAvailable(), test.WithToolsMcpSettings(test.McpServer()))},
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
	s.Run("Adds MCP clients for each MCP-enabled tools provider", func() {
		s.Require().Len(s.Ai.mcpClients, 1, "Expected MCP clients to be initialized")
	})
	s.Run("MCP clients are initialized", func() {
		s.False(slices.ContainsFunc(s.Ai.mcpClients, func(c *client.Client) bool { return !c.IsInitialized() }), "Expected all MCP clients to be initialized")
	})
}

func (s *AiMcpSuite) TestCloseShutsDownMcpServers() {
	s.Require().NotEmpty(s.Ai.mcpClients, "Expected MCP clients to be initialized")
	s.Ai.mcpClients[0].OnNotification(func(notification m3lmcp.JSONRPCNotification) {
		println(notification.Method)
	})
	s.Ai.Close()
	s.Run("MCP clients are shut down", func() {
		err := s.Ai.mcpClients[0].Ping(s.T().Context())
		s.Error(err, "Expected error when pinging closed MCP client")
		s.Contains(err.Error(), "closed", "Expected connection is closed error")
	})
}

func (s *AiMcpSuite) TestSetupAgentAddsMcpTools() {
	var receivedTools []*schema.ToolInfo
	s.Llm.WithToolsFunc = func(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
		receivedTools = tools
		return s.Llm, nil
	}
	s.Ai.Input <- api.NewUserMessage("Hello AItana! I'm sending some MCP tools.")
	s.Require().Eventually(func() bool { return len(receivedTools) > 0 }, 10*time.Second, 100, "Expected LLM to be called with tools")
	s.Run("Tools includes MCP Server tools", func() {
		s.True(slices.ContainsFunc(receivedTools, func(t *schema.ToolInfo) bool { return t.Name == "test-func" }), "Expected to find MCP 'test-func' tool in received tools")
	})
}

func TestAiMcp(t *testing.T) {
	suite.Run(t, new(AiMcpSuite))
}
