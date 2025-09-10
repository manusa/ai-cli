package test

import (
	"context"

	"github.com/manusa/ai-cli/pkg/api"
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
	Initialized bool        `json:"-"`
	Tools       []*api.Tool `json:"-"`
}

func (t *ToolsProvider) Initialize(_ context.Context) {
	t.Initialized = true
}

func (t *ToolsProvider) GetTools(_ context.Context) []*api.Tool {
	return t.Tools
}
