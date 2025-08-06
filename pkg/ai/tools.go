package ai

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/api"
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

var _ tool.InvokableTool = &invokableTool{}

func toType(t api.ToolParameterType) schema.DataType {
	switch t {
	case api.String:
		return schema.String
	}
	return schema.Object
}

func toInvokableTool(t *api.Tool) *invokableTool {
	var parameters *schema.ParamsOneOf
	if t.ParametersSchema != nil {
		parameters = schema.NewParamsOneOfByOpenAPIV3(t.ParametersSchema)
	} else {
		params := make(map[string]*schema.ParameterInfo, len(t.Parameters))
		for parameterKey, parameter := range t.Parameters {
			params[parameterKey] = &schema.ParameterInfo{
				Type:     toType(parameter.Type),
				Desc:     parameter.Description,
				Required: parameter.Required,
			}
		}
		parameters = schema.NewParamsOneOfByParams(params)
	}

	toolInfo := &schema.ToolInfo{
		Name:        t.Name,
		Desc:        t.Description,
		ParamsOneOf: parameters,
	}
	return &invokableTool{function: t.Function, toolInfo: toolInfo}
}
