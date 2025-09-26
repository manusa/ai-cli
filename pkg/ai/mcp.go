package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/eino-contrib/jsonschema"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ToolsProviderMcpClient struct {
	api.ToolsProvider
	*mcp.ClientSession
}

// Adaptation of https://github.com/cloudwego/eino-ext/blob/4a4306a8bf2cdae95b3e95bbe05b40fda0475fc2/components/tool/mcp/mcp.go
// to deal with https://github.com/cloudwego/eino-ext/issues/436
type mcpTool struct {
	toolInfo *schema.ToolInfo
	cli      *ToolsProviderMcpClient
}

var _ ToolManagerTool = &mcpTool{}

func (m *mcpTool) ToolsProvider() api.ToolsProvider {
	return m.cli.ToolsProvider
}

func (m *mcpTool) ToolInfo() *schema.ToolInfo {
	return m.toolInfo
}

func (m *mcpTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return m.toolInfo, nil
}

func (m *mcpTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	result, err := m.cli.CallTool(ctx, &mcp.CallToolParams{
		Name:      m.toolInfo.Name,
		Arguments: json.RawMessage(argumentsInJSON),
	})
	if err != nil {
		return "", fmt.Errorf("failed to call mcp tool: %w", err)
	}

	marshaledResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mcp tool result: %w", err)
	}
	return string(marshaledResult), nil
}

func toMcpTools(ctx context.Context, cli *ToolsProviderMcpClient) ([]ToolManagerTool, error) {
	listResults, err := cli.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("list mcp toolManager fail: %w", err)
	}

	ret := make([]ToolManagerTool, 0, len(listResults.Tools))
	for _, t := range listResults.Tools {
		marshaledInputSchema, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("conv mcp tool input schema fail(marshal): %w, tool name: %s", err, t.Name)
		}
		inputSchema := &jsonschema.Schema{}
		err = json.Unmarshal(marshaledInputSchema, inputSchema)
		if err != nil {
			return nil, fmt.Errorf("conv mcp tool input schema fail(unmarshal): %w, tool name: %s", err, t.Name)
		}

		ret = append(ret, &mcpTool{
			toolInfo: &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByJSONSchema(inputSchema),
			},
			cli: cli,
		})
	}

	return ret, nil
}

func ToMcpTools(ctx context.Context, mcpClients []*ToolsProviderMcpClient) (tools []ToolManagerTool) {
	tools = make([]ToolManagerTool, 0)
	for _, mcpClient := range mcpClients {
		mcpTools, err := toMcpTools(ctx, mcpClient)
		if err != nil {
			// TODO: log error
			continue
		}
		tools = append(tools, mcpTools...)
	}
	return tools
}

type HeaderRoundTripper struct {
	headers map[string]string
}

func (h *HeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func StartMcpClients(ctx context.Context, toolsProviders []api.ToolsProvider) []*ToolsProviderMcpClient {
	mcpClients := make([]*ToolsProviderMcpClient, 0, len(toolsProviders))
	for _, toolProvider := range toolsProviders {
		mcpSettings := toolProvider.GetMcpSettings()
		if mcpSettings == nil {
			continue
		}
		mcpClient := mcp.NewClient(&mcp.Implementation{Name: version.BinaryName + "-mcp-client", Version: version.Version}, nil)
		var tr mcp.Transport
		switch mcpSettings.Type {
		case api.McpTypeStdio:
			command := exec.CommandContext(ctx, mcpSettings.Command, mcpSettings.Args...)
			command.Env = append(os.Environ(), mcpSettings.Env...)
			tr = &mcp.CommandTransport{Command: command}
		case api.McpTypeSse:
			tr = &mcp.SSEClientTransport{
				Endpoint: mcpSettings.Url,
				HTTPClient: &http.Client{Transport: &HeaderRoundTripper{
					headers: mcpSettings.Headers,
				}},
			}
		case api.McpTypeStreamableHttp:
			tr = &mcp.StreamableClientTransport{
				Endpoint: mcpSettings.Url,
				HTTPClient: &http.Client{Transport: &HeaderRoundTripper{
					headers: mcpSettings.Headers,
				}},
			}
		}
		mcpSession, err := mcpClient.Connect(ctx, tr, nil)
		if err != nil {
			// TODO: log error
			continue
		}
		mcpClients = append(mcpClients, &ToolsProviderMcpClient{ToolsProvider: toolProvider, ClientSession: mcpSession})
	}
	return mcpClients
}

func StopMcpClients(mcpClients []*ToolsProviderMcpClient) {
	for _, mcpClient := range mcpClients {
		_ = mcpClient.Close()
	}
}
