package inference

import (
	"fmt"
	"slices"
	"strings"

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

// Discover the available and not available inference providers based on the user preferences
func Discover(cfg *config.Config, policies map[string]any) (availableInferences []api.InferenceProvider, notAvailableInferences []api.InferenceProvider) {
	availableInferences, notAvailableInferences = []api.InferenceProvider{}, []api.InferenceProvider{}
	for _, provider := range providers {
		if provider.IsAvailable(cfg, policies) {
			availableInferences = append(availableInferences, provider)
		} else {
			notAvailableInferences = append(notAvailableInferences, provider)
		}
	}
	slices.SortFunc(availableInferences, func(a, b api.InferenceProvider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	slices.SortFunc(notAvailableInferences, func(a, b api.InferenceProvider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return availableInferences, notAvailableInferences
}
