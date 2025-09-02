package cursor

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
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

func (s *CursorTestSuite) SetupTest() {}

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

func (s *CursorTestSuite) TestGetConfigWithTools() {
	s.Run("GetConfig returns a config with the tools", func() {
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
		s.JSONEq(string(result), `{ "mcpServers": { "testtool": { "command": "mycmd", "args": ["arg1", "arg2"], "env": { "ENV": "value" } } } }`)
	})
}
