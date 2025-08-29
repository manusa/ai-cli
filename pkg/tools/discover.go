package tools

import (
	"fmt"
	"maps"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
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
func Initialize(cfg *config.Config, policies map[string]any) []api.ToolsProvider {
	for _, provider := range providers {
		provider.Initialize(cfg, policies[provider.Attributes().Name()])
	}
	return slices.SortedFunc(maps.Values(providers), api.FeatureSorter)
}

func GetDefaultPolicies() map[string]any {
	policies := make(map[string]any)
	for _, provider := range providers {
		providerPolicies := provider.GetDefaultPolicies()
		policies[provider.Attributes().Name()] = providerPolicies
	}
	return policies
}
