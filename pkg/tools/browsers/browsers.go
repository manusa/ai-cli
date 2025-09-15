package browsers

import (
	"context"

	browserslib "github.com/feloy/browsers-mcp-server/pkg/browsers"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

var (
	supportedMcpSettings = api.McpSettings{
		Type:    api.McpTypeStdio,
		Command: "npx",
		Args:    []string{"-y", "browsers-mcp-server@latest"},
	}
)

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.ToolsParameters = cfg.ToolsParameters(p.Attributes().Name())
	}

	detectedBrowsers := browserslib.GetBrowsers()
	if len(detectedBrowsers) == 0 {
		p.IsAvailableReason = "no browsers detected"
		return
	}

	var err error
	p.McpSettings, err = p.findBestMcpServerSettings()
	if err != nil {
		p.IsAvailableReason = err.Error()
		return
	}

	p.Available = true
	p.IsAvailableReason = "always available"
}

func (p *Provider) findBestMcpServerSettings() (*api.McpSettings, error) {
	return &supportedMcpSettings, nil
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "browsers",
				FeatureDescription: "Provides access to browsers",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
