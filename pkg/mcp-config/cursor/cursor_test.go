package cursor

import (
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type CursorTestSuite struct {
	suite.Suite
}

type TestToolsProvider struct {
	test.ToolsProvider
	mcpSettings *api.McpSettings
}

func (p *TestToolsProvider) GetMcpSettings() *api.McpSettings {
	return p.mcpSettings
}

func (s *CursorTestSuite) SetupTest() {
	config.FileSystem = afero.NewMemMapFs()
	xdg.Home = "/path/to/home"
}

func TestCursor(t *testing.T) {
	suite.Run(t, new(CursorTestSuite))
}

func (s *CursorTestSuite) TestGetConfigEmpty() {
	s.Run("GetConfig returns an empty config with no tools", func() {
		provider := &CursorMcpConfig{}
		var tools []api.ToolsProvider
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": {} }`)
	})
}

func (s *CursorTestSuite) TestGetConfigWithToolsWhenNoConfigExists() {
	s.Run("GetConfig returns a config with the tools when no config exists", func() {
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{
			&TestToolsProvider{
				ToolsProvider: test.ToolsProvider{
					BasicToolsProvider: api.BasicToolsProvider{
						BasicToolsAttributes: api.BasicToolsAttributes{
							BasicFeatureAttributes: api.BasicFeatureAttributes{
								FeatureName:        "testtool",
								FeatureDescription: "Test Tool",
							},
						},
					},
				},
				mcpSettings: &api.McpSettings{
					Type:    api.McpTypeStdio,
					Command: "mycmd",
					Args:    []string{"arg1", "arg2"},
					Env:     []string{"ENV=value"},
				},
			},
		}
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": { "ai-cli-testtool": { "type": "stdio", "command": "mycmd", "args": ["arg1", "arg2"], "env": { "ENV": "value" } } } }`)
	})
}

func (s *CursorTestSuite) TestGetConfigAddToolsWhenConfigExists() {
	s.Run("GetConfig returns a config with the tools added when a config exists", func() {
		err := createFile(config.FileSystem, "/path/to/home/.cursor/mcp.json", `
{
	"mcpServers": {
		"other-tool": {
			"type": "stdio",
			"command": "other-tool",
			"args": [],
			"env": {
				"ENV1": "value1"
			}
		}
	}
}`)
		s.NoError(err)
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{
			&TestToolsProvider{
				ToolsProvider: test.ToolsProvider{
					BasicToolsProvider: api.BasicToolsProvider{
						BasicToolsAttributes: api.BasicToolsAttributes{
							BasicFeatureAttributes: api.BasicFeatureAttributes{
								FeatureName:        "testtool",
								FeatureDescription: "Test Tool",
							},
						},
					},
				},
				mcpSettings: &api.McpSettings{
					Type:    api.McpTypeStdio,
					Command: "mycmd",
					Args:    []string{"arg1", "arg2"},
					Env:     []string{"ENV=value"},
				},
			},
		}
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": { "other-tool": { "type": "stdio", "command": "other-tool", "args": [], "env": { "ENV1": "value1" } }, "ai-cli-testtool": { "type": "stdio", "command": "mycmd", "args": ["arg1", "arg2"], "env": { "ENV": "value" } } } }`)
	})
}

func (s *CursorTestSuite) TestGetConfigUpdateToolsWhenConfigExists() {
	s.Run("GetConfig returns a config with the tools updated when a config exists", func() {
		err := createFile(config.FileSystem, "/path/to/home/.cursor/mcp.json", `
{
	"mcpServers": {
		"other-tool": {
			"type": "stdio",
			"command": "other-tool",
			"args": [],
			"env": {
				"ENV1": "value1"
			}
		},
		"ai-cli-testtool": {
			"type": "stdio",
			"command": "my-previous-cmd",
			"args": ["arg1", "arg2", "arg3"],
			"env": {
				"ENV": "previous-value"
			}
		}
	}
}`)
		s.NoError(err)
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{
			&TestToolsProvider{
				ToolsProvider: test.ToolsProvider{
					BasicToolsProvider: api.BasicToolsProvider{
						BasicToolsAttributes: api.BasicToolsAttributes{
							BasicFeatureAttributes: api.BasicFeatureAttributes{
								FeatureName:        "testtool",
								FeatureDescription: "Test Tool",
							},
						},
					},
				},
				mcpSettings: &api.McpSettings{
					Type:    api.McpTypeStdio,
					Command: "mycmd",
					Args:    []string{"arg1", "arg2"},
					Env:     []string{"ENV=value"},
				},
			},
		}
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": { "other-tool": { "type": "stdio", "command": "other-tool", "args": [], "env": { "ENV1": "value1" } }, "ai-cli-testtool": { "type": "stdio", "command": "mycmd", "args": ["arg1", "arg2"], "env": { "ENV": "value" } } } }`)
	})
}

func (s *CursorTestSuite) TestGetConfigDeleteToolsWhenConfigExists() {
	s.Run("GetConfig returns a config with the tools deleted when a config exists", func() {
		err := createFile(config.FileSystem, "/path/to/home/.cursor/mcp.json", `
{
	"mcpServers": {
		"other-tool": {
			"type": "stdio",
			"command": "other-tool",
			"args": [],
			"env": {
				"ENV1": "value1"
			}
		},
		"ai-cli-testtool": {
			"type": "stdio",
			"command": "my-previous-cmd",
			"args": ["arg1", "arg2", "arg3"],
			"env": {
				"ENV": "previous-value"
			}
		}
	}
}`)
		s.NoError(err)
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{}
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": { "other-tool": { "type": "stdio", "command": "other-tool", "args": [], "env": { "ENV1": "value1" } } } }`)
	})
}

func createFile(fs afero.Fs, path string, content string) error {
	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write([]byte(content))
	return err
}
