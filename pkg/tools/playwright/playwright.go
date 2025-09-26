package playwright

import (
	"context"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
)

// TODO: Centralize version management elsewhere
const version = "0.0.40"

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.ToolsParameters = cfg.ToolsParameters(p.Attributes().Name())
	}
	if !config.CommandExists("npx") {
		p.IsAvailableReason = "npx command not found"
		return
	}
	p.IsAvailableReason = "npx command found"
	p.Available = true
	p.McpSettings = &api.McpSettings{}
	p.McpSettings.Type = api.McpTypeStdio
	p.McpSettings.Command = "npx"
	p.McpSettings.Args = []string{"-y", "@playwright/mcp@" + version}

	if !config.IsDesktop() {
		p.McpSettings.Args = append(p.McpSettings.Args, "--headless")
	}
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName: "playwright",
				FeatureDescription: "Enables web browsing capabilities through Playwright. " +
					"Opening web pages, opening URLs, interacting with elements inside the browser, extracting snapshots, and scraping information from web pages. " +
					"Support for multiple tabs and many other browser options",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
