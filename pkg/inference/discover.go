package inference

import (
	"fmt"
	"maps"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]api.InferenceProvider{}

// Register a new inference provider
func Register(provider api.InferenceProvider) {
	if provider == nil {
		panic("cannot register a nil inference provider")
	}
	if _, ok := providers[provider.Attributes().Name()]; ok {
		panic(fmt.Sprintf("inference provider already registered: %s", provider.Attributes().Name()))
	}
	providers[provider.Attributes().Name()] = provider
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]api.InferenceProvider{}
}

// Initialize initializes the registered providers based on the user preferences
func Initialize(cfg *config.Config, policies map[string]any) []api.InferenceProvider {
	for _, provider := range providers {
		provider.Initialize(cfg, policies[provider.Attributes().Name()])
	}
	return slices.Collect(maps.Values(providers))
}
