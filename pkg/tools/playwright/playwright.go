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
	api.McpSettings `json:"-"`
}

var _ api.ToolsProvider = &Provider{}

func (p *Provider) Initialize(_ context.Context) {
	if !config.CommandExists("npx") {
		p.IsAvailableReason = "npx command not found"
		return
	}
	p.IsAvailableReason = "npx command found"
	p.Available = true
	p.Command = "npx"
	p.Args = []string{"-y", "@playwright/mcp@" + version}

	if !config.IsDesktop() {
		p.Args = append(p.Args, "--headless")
	}
}

func (p *Provider) GetTools(ctx context.Context) ([]*api.Tool, error) {
	cli, err := eino.StartMcp(ctx, p.Env, slices.Concat([]string{p.Command}, p.Args))
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
