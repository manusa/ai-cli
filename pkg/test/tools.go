package test

import (
	"context"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

type ToolsProviderOption func(*ToolsProvider)

func WithToolsAvailable() ToolsProviderOption {
	return func(i *ToolsProvider) {
		i.Available = true
	}
}

func NewToolsProvider(name string, options ...ToolsProviderOption) *ToolsProvider {
	p := &ToolsProvider{
		BasicToolsProvider: api.BasicToolsProvider{
			BasicToolsAttributes: api.BasicToolsAttributes{
				BasicFeatureAttributes: api.BasicFeatureAttributes{
					FeatureName: name,
				},
			},
		},
	}
	for _, option := range options {
		option(p)
	}
	return p
}

type ToolsProvider struct {
	api.BasicToolsProvider
	Initialized bool           `json:"-"`
	Tools       []*api.Tool    `json:"-"`
	Policies    map[string]any `json:"-"`
}

func (t *ToolsProvider) Initialize(_ *config.Config, _ any) {
	t.Initialized = true
}

func (t *ToolsProvider) GetDefaultPolicies() map[string]any {
	return t.Policies
}

func (t *ToolsProvider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return t.Tools, nil
}
