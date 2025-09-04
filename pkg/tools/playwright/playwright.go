package playwright

import (
	"context"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

// TODO: Centralize version management elsewhere
const version = "0.0.36"

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

func (p *Provider) GetTools(ctx context.Context) ([]*api.Tool, error) {
	cli, err := eino.StartMcp(ctx, p.McpSettings.Env, slices.Concat([]string{p.McpSettings.Command}, p.McpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "playwright",
				FeatureDescription: "Automate and interact with web browsers using Playwright.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
