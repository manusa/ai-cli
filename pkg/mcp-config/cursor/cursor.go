package cursor

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/manusa/ai-cli/pkg/api"
)

type CursorMcpConfigFile struct {
	McpServers map[string]any `json:"mcpServers"`
}

type StdioServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type RemoteServerConfig struct {
	Url string `json:"url,omitempty"`
	/**
	 * Optional HTTP headers to include with every request to this server (e.g. for authentication).
	 * The keys are header names and the values are header values.
	 */
	Headers map[string]string `json:"headers,omitempty"`
}

type CursorMcpConfig struct{}

func (p *CursorMcpConfig) GetFile() string {
	return path.Join(xdg.Home, ".cursor", "mcp.json")
}

func (p *CursorMcpConfig) GetConfig(tools []api.ToolsProvider) ([]byte, error) {
	result := CursorMcpConfigFile{
		McpServers: make(map[string]any),
	}
	for _, tool := range tools {
		mcpSettings := tool.GetMcpSettings()
		if mcpSettings == nil {
			continue
		}
		if mcpSettings.Type == api.McpTypeStdio {
			result.McpServers[tool.Attributes().Name()] = StdioServerConfig{
				Command: mcpSettings.Command,
				Args:    mcpSettings.Args,
				Env:     toEnvMap(mcpSettings.Env),
			}
		}
	}
	return json.MarshalIndent(result, "", "  ")
}

func toEnvMap(envArray []string) map[string]string {
	result := make(map[string]string, len(envArray))
	for _, env := range envArray {
		key, value, _ := strings.Cut(env, "=")
		result[key] = value
	}
	return result
}
