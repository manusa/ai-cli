package tools

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
)

var providers = map[string]api.ToolsProvider{}

// Register a new tools provider
func Register(provider api.ToolsProvider) {
	if provider == nil {
		panic("cannot register a nil tools provider")
	}
	if _, ok := providers[provider.Attributes().Name()]; ok {
		panic(fmt.Sprintf("tool provider already registered: %s", provider.Attributes().Name()))
	}
	providers[provider.Attributes().Name()] = provider
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]api.ToolsProvider{}
}

// Initialize initializes the registered providers based on the user preferences
func Initialize(ctx context.Context) []api.ToolsProvider {
	for _, provider := range providers {
		provider.Initialize(ctx)
	}
	return slices.SortedFunc(maps.Values(providers), api.FeatureSorter)
}

func GetProviders() map[string]api.ToolsProvider {
	return providers
}

// GetDefaultParameters returns the default tool parameters for the registered providers
// TODO: all parameters are set to false by default, do we want to change this?
func GetDefaultParameters() map[string]api.ToolParameters {
	cfg := map[string]api.ToolParameters{}
	for _, provider := range providers {
		cfg[provider.Attributes().Name()] = api.ToolParameters{
			Local:          false,
			NonDestructive: false,
			ReadOnly:       false,
		}
	}
	return cfg
}
