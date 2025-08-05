package eino

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/mark3labs/mcp-go/client"
	m3lmcp "github.com/mark3labs/mcp-go/mcp"
)

func StartMcp(ctx context.Context, cmdAndArgs []string) (*client.Client, error) {
	cli, err := client.NewStdioMCPClient(cmdAndArgs[0], []string{}, cmdAndArgs[1:]...)
	if err != nil {
		return nil, err
	}
	initRequest := m3lmcp.InitializeRequest{}
	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func GetTools(ctx context.Context, cli client.MCPClient) ([]*api.Tool, error) {
	tools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
	if err != nil {
		return nil, err
	}
	apiTools := make([]*api.Tool, len(tools))
	for i, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			return nil, err
		}
		schema, err := info.ToOpenAPIV3()
		if err != nil {
			return nil, err
		}
		apiTools[i] = &api.Tool{
			Name:             info.Name,
			Description:      info.Desc,
			ParametersSchema: schema,
			Function: func(args map[string]interface{}) (string, error) {
				jsonArgs, err := json.Marshal(args)
				if err != nil {
					return "", err
				}
				return tool.(einotool.InvokableTool).InvokableRun(ctx, string(jsonArgs))
			},
		}
	}

	return apiTools, nil
}
