package ai

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ToolManager struct {
	availableTools map[string]tool.InvokableTool
	enabledTools   map[string]tool.InvokableTool
	enableTool     *invokableTool
}

func NewToolManager(ctx context.Context, availableTools []tool.InvokableTool) *ToolManager {
	availableToolsMap := make(map[string]tool.InvokableTool, len(availableTools))
	toolNameParameter := strings.Builder{}
	toolNameParameter.WriteString("The name of the tool or tools to enable.\n")
	toolNameParameter.WriteString("You can enable multiple tools by separating their names with commas.\n")
	toolNameParameter.WriteString("You can pick the tools from the following list (xml):\n<tools>\n")
	for _, t := range availableTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		availableToolsMap[info.Name] = t
		toolNameParameter.WriteString(fmt.Sprintf(`<tool name="%s">%s</tool>`, info.Name, info.Desc) + "\n")
	}
	toolNameParameter.WriteString("\n</tools>")
	toolManager := &ToolManager{
		availableTools: availableToolsMap,
		enabledTools:   make(map[string]tool.InvokableTool),
	}
	toolManager.enableTool = &invokableTool{
		toolInfo: &schema.ToolInfo{
			Name: "tool_enable",
			Desc: "Enable a tool for the current session.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"tool_names": {
					Type:     schema.String,
					Desc:     toolNameParameter.String(),
					Required: true,
					Enum:     slices.Collect(maps.Keys(toolManager.availableTools)),
				},
			}),
		},
		function: toolManager.toolEnable,
	}
	toolManager.availableTools[toolManager.enableTool.toolInfo.Name] = toolManager.enableTool
	return toolManager
}

func (t *ToolManager) ToolCount() int {
	return len(t.enabledTools)
}

func (t *ToolManager) EnabledToolsReset() {
	t.enabledTools = make(map[string]tool.InvokableTool)
}

func (t *ToolManager) EnabledTools() []tool.BaseTool {
	ret := make([]tool.BaseTool, 0, len(t.enabledTools))
	for _, enabledTool := range t.enabledTools {
		ret = append(ret, enabledTool)
	}
	ret = append(ret, t.enableTool) // Always include the enable tool
	return ret
}

func (t *ToolManager) InvokeTool(ctx context.Context, name, input string) (string, error) {
	if enabledTool, exists := t.enabledTools[name]; exists {
		return enabledTool.InvokableRun(ctx, input)
	}
	if _, exists := t.availableTools[name]; exists {
		return fmt.Sprintf("Tool '%s' is not enabled. You can enable it by calling the 'tool_enable' tool first.", name), nil
	}
	return fmt.Sprintf("Tool '%s' not found.", name), nil
}

func (t *ToolManager) toolEnable(args map[string]interface{}) (string, error) {
	toolNames, ok := args["tool_names"].(string)
	if !ok {
		return "Invalid tool names.", nil
	}
	sb := strings.Builder{}
	for _, toolName := range strings.Split(toolNames, ",") {
		toolName = strings.TrimSpace(toolName)
		if _, exists := t.enabledTools[toolName]; exists {
			sb.WriteString(fmt.Sprintf("Tool '%s' was already enabled.", toolName))
			continue
		}
		if _, exists := t.availableTools[toolName]; !exists {
			sb.WriteString(fmt.Sprintf("Tool '%s' not found.", toolName))
			continue
		}
		t.enabledTools[toolName] = t.availableTools[toolName]
		sb.WriteString(fmt.Sprintf("Tool '%s' enabled.", toolName))
	}
	return sb.String(), nil
}
