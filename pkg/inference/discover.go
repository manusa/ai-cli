package inference

import (
	"context"
	"fmt"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/policies"
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

// Initialize initializes the registered providers based on the user preferences and policies
func Initialize(ctx context.Context) (disabled []api.InferenceProvider, enabled []api.InferenceProvider) {
	ctxPolicies := policies.GetPolicies(ctx)
	for _, provider := range providers {
		if ctxPolicies != nil && !policies.PoliciesProvider.IsInferenceEnabledByPolicies(provider, ctxPolicies) {
			disabled = append(disabled, provider)
			continue
		}
		provider.Initialize(ctx)
		enabled = append(enabled, provider)
	}
	return disabled, enabled
}
