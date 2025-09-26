package ai

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/api"
)

type invokableTool struct {
	toolsProvider api.ToolsProvider
	toolInfo      *schema.ToolInfo
	function      func(args map[string]interface{}) (string, error)
}

var _ ToolManagerTool = &invokableTool{}

func (i invokableTool) ToolsProvider() api.ToolsProvider {
	return i.toolsProvider
}

func (i invokableTool) ToolInfo() *schema.ToolInfo {
	return i.toolInfo
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

func toType(t api.ToolParameterType) schema.DataType {
	switch t {
	case api.String:
		return schema.String
	}
	return schema.Object
}

func toInvokableTools(ctx context.Context, toolsProviders []api.ToolsProvider) (tools []ToolManagerTool) {
	tools = make([]ToolManagerTool, 0)
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
				Name:        provider.Attributes().Name() + "_" + t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByParams(params),
			}
			tools = append(tools, &invokableTool{function: t.Function, toolInfo: toolInfo, toolsProvider: provider})
		}
	}
	return tools
}
