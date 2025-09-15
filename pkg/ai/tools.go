package ai

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	m3lmcp "github.com/mark3labs/mcp-go/mcp"
)

type invokableTool struct {
	toolInfo *schema.ToolInfo
	function func(args map[string]interface{}) (string, error)
}

func (i invokableTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return i.toolInfo, nil
}

func (i invokableTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args map[string]interface{}
	if argumentsInJSON != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return "", err
		}

	}
	return i.function(args)
}

var _ tool.BaseTool = &invokableTool{}
var _ tool.InvokableTool = &invokableTool{}

func toType(t api.ToolParameterType) schema.DataType {
	switch t {
	case api.String:
		return schema.String
	}
	return schema.Object
}

func toInvokableTools(ctx context.Context, toolsProviders []api.ToolsProvider) (tools []tool.BaseTool) {
	tools = make([]tool.BaseTool, 0)
	for _, provider := range toolsProviders {
		for _, t := range provider.GetTools(ctx) {
			params := make(map[string]*schema.ParameterInfo, len(t.Parameters))
			for parameterKey, parameter := range t.Parameters {
				params[parameterKey] = &schema.ParameterInfo{
					Type:     toType(parameter.Type),
					Desc:     parameter.Description,
					Required: parameter.Required,
				}
			}

			toolInfo := &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByParams(params),
			}
			tools = append(tools, &invokableTool{function: t.Function, toolInfo: toolInfo})
		}
	}
	return tools
}

func startMcpClients(ctx context.Context, toolsProviders []api.ToolsProvider) []*client.Client {
	mcpClients := make([]*client.Client, 0, len(toolsProviders))
	for _, tool := range toolsProviders {
		mcpSettings := tool.GetMcpSettings()
		if mcpSettings == nil {
			continue
		}
		var mcpClient *client.Client
		var err error
		switch mcpSettings.Type {
		case api.McpTypeStdio:
			mcpClient, err = client.NewStdioMCPClient(mcpSettings.Command, mcpSettings.Env, mcpSettings.Args...)
		case api.McpTypeSse:
			mcpClient, err = client.NewSSEMCPClient(mcpSettings.Url, client.WithHeaders(mcpSettings.Headers))
		case api.McpTypeStreamableHttp:
			mcpClient, err = client.NewStreamableHttpClient(mcpSettings.Url, transport.WithHTTPHeaders(mcpSettings.Headers))
		}
		if err != nil {
			// TODO: log error
			continue
		}
		initRequest := m3lmcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = m3lmcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = m3lmcp.Implementation{Name: version.BinaryName, Version: version.Version}
		_, err = mcpClient.Initialize(ctx, initRequest)
		if err != nil {
			// TODO: log error
			continue
		}
		mcpClients = append(mcpClients, mcpClient)
	}
	return mcpClients
}

func stopMcpClients(mcpClients []*client.Client) {
	for _, mcpClient := range mcpClients {
		_ = mcpClient.Close()
	}
}

func mcpClientTools(ctx context.Context, mcpClients []*client.Client) (tools []tool.BaseTool) {
	tools = make([]tool.BaseTool, 0)
	for _, mcpClient := range mcpClients {
		baseTools, err := mcp.GetTools(ctx, &mcp.Config{
			Cli: mcpClient,
			ToolCallResultHandler: func(ctx context.Context, name string, result *m3lmcp.CallToolResult) (*m3lmcp.CallToolResult, error) {
				// https://github.com/cloudwego/eino-ext/issues/436
				if result.IsError {
					if result.Meta == nil {
						result.Meta = &m3lmcp.Meta{AdditionalFields: make(map[string]interface{})}
					}
					result.Meta.AdditionalFields["error"] = true
					result.IsError = false
				}
				return result, nil
			},
		})
		if err != nil {
			// TODO: log error
			continue
		}
		tools = append(tools, baseTools...)
	}
	return tools
}
