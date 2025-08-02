package kubernetes

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/mark3labs/mcp-go/client"
	mmcp "github.com/mark3labs/mcp-go/mcp"
	"k8s.io/client-go/tools/clientcmd"
)

type Provider struct {
}

var _ tools.Provider = &Provider{}

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "kubernetes",
		},
	}
}

func (p *Provider) IsAvailable(_ *config.Config) bool {
	_, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	return err == nil
}

func (p *Provider) GetTools(ctx context.Context, _ *config.Config) ([]*api.Tool, error) {
	// TODO: start the MCP Server
	cli, err := client.NewSSEMCPClient("http://localhost:8080/sse")
	if err != nil {
		return nil, err
	}
	err = cli.Start(ctx)
	if err != nil {
		return nil, err
	}
	initRequest := mmcp.InitializeRequest{}
	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
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
		parameters, err := toToolParameters(info)
		if err != nil {
			return nil, err
		}
		apiTools[i] = &api.Tool{
			Name:        info.Name,
			Description: info.Desc,
			Parameters:  parameters,
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

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Attributes())
}

func toToolParameters(info *schema.ToolInfo) (map[string]api.ToolParameter, error) {
	schema, err := info.ToOpenAPIV3()
	if err != nil {
		return nil, err
	}
	parameters := make(map[string]api.ToolParameter)
	for k, property := range schema.Properties {
		parameters[k] = api.ToolParameter{
			Required:    isRequired(schema.Required, k),
			Type:        api.ToolParameterType(property.Value.Type),
			Description: property.Value.Description,
		}
	}
	return parameters, nil
}

func isRequired(required []string, k string) bool {
	for _, r := range required {
		if r == k {
			return true
		}
	}
	return false
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
