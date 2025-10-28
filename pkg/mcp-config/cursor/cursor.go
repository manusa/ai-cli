package cursor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

const prefix = "ai-cli-"

type CursorMcpConfigFile struct {
	McpServers map[string]any `json:"mcpServers"`
}

type StdioServerConfig struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	EnvFile string            `json:"envFile,omitempty"`
}

type RemoteServerConfig struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url,omitempty"`
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
	configFile := p.GetFile()
	var existingConfig *CursorMcpConfigFile
	if _, err := config.FileSystem.Stat(configFile); err == nil {
		existingConfig, err = readExistingConfig(configFile)
		if err != nil {
			return nil, err
		}
	}
	if existingConfig == nil {
		existingConfig = &CursorMcpConfigFile{
			McpServers: make(map[string]any),
		}
	}
	if existingConfig.McpServers == nil {
		existingConfig.McpServers = make(map[string]any)
	}

	// Remove existing tools prefixed with our prefix
	for name := range existingConfig.McpServers {
		if strings.HasPrefix(name, prefix) {
			delete(existingConfig.McpServers, name)
		}
	}

	// Add our tools with our prefix
	for _, tool := range tools {
		mcpSettings := tool.GetMcpSettings()
		if mcpSettings == nil {
			continue
		}
		switch mcpSettings.Type {
		case api.McpTypeStdio:
			existingConfig.McpServers[toolName(tool)] = StdioServerConfig{
				Type:    mcpSettings.Type.String(),
				Command: mcpSettings.Command,
				Args:    mcpSettings.Args,
				Env:     toEnvMap(mcpSettings.Env),
			}
		case api.McpTypeStreamableHttp:
			existingConfig.McpServers[toolName(tool)] = RemoteServerConfig{
				Type:    mcpSettings.Type.String(),
				Url:     mcpSettings.Url,
				Headers: mcpSettings.Headers,
			}
		default:
			continue
		}
	}
	return json.MarshalIndent(existingConfig, "", "  ")
}

func readExistingConfig(configFile string) (*CursorMcpConfigFile, error) {
	file, err := config.FileSystem.OpenFile(configFile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var existingConfig CursorMcpConfigFile
	err = json.Unmarshal(fileContent, &existingConfig)
	if err != nil {
		return nil, err
	}
	return &existingConfig, nil
}

func toEnvMap(envArray []string) map[string]string {
	result := make(map[string]string, len(envArray))
	for _, env := range envArray {
		key, value, _ := strings.Cut(env, "=")
		result[key] = value
	}
	return result
}

func toolName(tool api.ToolsProvider) string {
	return fmt.Sprintf("%s%s", prefix, tool.Attributes().Name())
}
