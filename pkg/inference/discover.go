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

func GetProviders() map[string]api.InferenceProvider {
	return providers
}

func Unregister(name string) {
	delete(providers, name)
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]api.InferenceProvider{}
}

// Initialize initializes the registered providers based on the user preferences
func Initialize(ctx context.Context) []api.InferenceProvider {
	for _, provider := range providers {
		provider.Initialize(ctx)
	}
	return slices.Collect(maps.Values(providers))
}
