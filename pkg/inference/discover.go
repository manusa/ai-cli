package inference

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
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
func Initialize(ctx context.Context, policies map[string]any) []api.InferenceProvider {
	for _, provider := range providers {
		provider.Initialize(ctx, policies[provider.Attributes().Name()])
	}
	return slices.Collect(maps.Values(providers))
}
