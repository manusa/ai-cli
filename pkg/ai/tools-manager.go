package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/manusa/ai-cli/pkg/api"
)

type ToolManagerTool interface {
	tool.InvokableTool
	tool.BaseTool
	ToolsProvider() api.ToolsProvider
	ToolInfo() *schema.ToolInfo
}

type ToolManager struct {
	availableTools []ToolManagerTool
	enabledTools   map[string]ToolManagerTool
	enableTool     ToolManagerTool
}

func NewToolManager(toolsProviders []api.ToolsProvider, availableTools []ToolManagerTool) *ToolManager {
	toolsetNames := make([]string, 0, len(toolsProviders))
	toolNameParameter := strings.Builder{}
	toolNameParameter.WriteString("The name or names of the toolsets to enable.\n")
	toolNameParameter.WriteString("You can enable multiple toolsets separating their names with commas.\n")
	toolNameParameter.WriteString("You can pick the toolsets from the following list (xml):\n")
	toolNameParameter.WriteString("<toolsets>\n")
	for _, t := range toolsProviders {
		toolsetNames = append(toolsetNames, t.Attributes().Name())
		toolNameParameter.WriteString(fmt.Sprintf(`<toolset name="%s">%s</tool>`, t.Attributes().Name(), t.Attributes().Description()) + "\n")
	}
	toolNameParameter.WriteString("\n</toolsets>")
	toolManager := &ToolManager{
		availableTools: availableTools,
		enabledTools:   make(map[string]ToolManagerTool),
	}
	toolManager.enableTool = &invokableTool{
		toolInfo: &schema.ToolInfo{
			Name: "toolset_enable",
			Desc: "Enable a toolset for the current session.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"toolset_names": {
					Type:     schema.String,
					Desc:     toolNameParameter.String(),
					Required: true,
					Enum:     toolsetNames,
				},
			}),
		},
		function: toolManager.toolsetEnable,
	}
	return toolManager
}

func (t *ToolManager) ToolEnabledCount() int {
	return len(t.enabledTools)
}

func (t *ToolManager) ToolCount() int {
	return len(t.availableTools)
}

func (t *ToolManager) EnabledToolsReset() {
	t.enabledTools = make(map[string]ToolManagerTool)
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
	// TODO: Report the toolset the tool belongs to, and suggest enabling it
	//if _, exists := t.availableTools[name]; exists {
	//	return fmt.Sprintf("Tool '%s' is not enabled. You can enable it by calling the 'tool_enable' tool first.", name), nil
	//}
	return fmt.Sprintf("Tool '%s' not found.", name), nil
}

func (t *ToolManager) toolsetEnable(args map[string]interface{}) (string, error) {
	toolsetNames, ok := args["toolset_names"].(string)
	if !ok {
		return "Invalid toolset names.", nil
	}
	sb := strings.Builder{}
	for _, toolsetName := range strings.Split(toolsetNames, ",") {
		toolsetName = strings.TrimSpace(toolsetName)
		for _, availableTool := range t.availableTools {
			// The built-in enable tool does not have a provider
			if availableTool.ToolsProvider() != nil && availableTool.ToolsProvider().Attributes().Name() != toolsetName {
				continue
			}
			toolName := availableTool.ToolInfo().Name
			if _, exists := t.enabledTools[toolName]; exists {
				// TODO: probably not necessary
				//sb.WriteString(fmt.Sprintf("Tool '%s' was already enabled.", toolName))
				continue
			}
			t.enabledTools[toolName] = availableTool
		}
		sb.WriteString(fmt.Sprintf("Toolset '%s' enabled.", toolsetName))
	}
	return sb.String(), nil
}
