package cursor

import (
	"context"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/stretchr/testify/suite"
)

type CursorTestSuite struct {
	suite.Suite
}

type TestToolsProvider struct {
	mcpSettings *api.McpSettings
}

func (p *TestToolsProvider) GetMcpSettings() *api.McpSettings {
	return p.mcpSettings
}

func (p *TestToolsProvider) Attributes() api.ToolsAttributes {
	return &tools.BasicToolsProvider{
		BasicToolsAttributes: tools.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "testtool",
				FeatureDescription: "A test tool",
			},
		},
	}
}

func (p *TestToolsProvider) GetDefaultPolicies() map[string]any {
	return nil
}

func (p *TestToolsProvider) GetTools(ctx context.Context, cfg *config.Config) ([]*api.Tool, error) {
	return nil, nil
}

func (p *TestToolsProvider) IsAvailable(_ *config.Config, _ any) bool {
	return true
}

func (p *TestToolsProvider) Reason() string {
	return ""
}

func (s *CursorTestSuite) SetupTest() {}

func TestCursor(t *testing.T) {
	suite.Run(t, new(CursorTestSuite))
}

func (s *CursorTestSuite) TestGetConfigEmpty() {
	s.Run("GetConfig returns an empty config with no tools", func() {
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{}
		result, err := provider.GetConfig(tools)
		s.NoError(err)
		s.JSONEq(string(result), `{ "mcpServers": {} }`)
	})
}

func (s *CursorTestSuite) TestGetConfigWithTools() {
	s.Run("GetConfig returns a config with the tools", func() {
		provider := &CursorMcpConfig{}
		tools := []api.ToolsProvider{
			&TestToolsProvider{
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
		s.JSONEq(string(result), `{ "mcpServers": { "testtool": { "command": "mycmd", "args": ["arg1", "arg2"], "env": { "ENV": "value" } } } }`)
	})
}
