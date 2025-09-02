package tools

import (
	"context"
	"fmt"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/policies"
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
func Initialize(ctx context.Context) (disabled []api.ToolsProvider, enabled []api.ToolsProvider) {
	ctxPolicies := policies.GetPolicies(ctx)
	for _, provider := range providers {
		if ctxPolicies != nil && !policies.PoliciesProvider.IsToolEnabledByPolicies(provider, ctxPolicies) {
			disabled = append(disabled, provider)
			continue
		}
		provider.Initialize(ctx)
		enabled = append(enabled, provider)
	}
	return disabled, enabled
}
