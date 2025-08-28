package tools

import (
	"fmt"
	"slices"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]api.ToolsProvider{}

type BasicToolsProvider struct {
	api.ToolsProvider `json:"-"`
	BasicToolsAttributes
	IsAvailableReason string           `json:"reason"`
	McpSettings       *api.McpSettings `json:"mcp_settings,omitempty"`
}

func (p *BasicToolsProvider) Attributes() api.ToolsAttributes {
	return &p.BasicToolsAttributes
}

func (p *BasicToolsProvider) Reason() string {
	return p.IsAvailableReason
}

type BasicToolsAttributes struct {
	api.BasicFeatureAttributes
}

// Register a new tools provider
func Register(provider api.ToolsProvider) {
	if _, ok := providers[provider.Attributes().Name()]; ok {
		panic(fmt.Sprintf("tool provider already registered: %s", provider.Attributes().Name()))
	}
	providers[provider.Attributes().Name()] = provider
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]api.ToolsProvider{}
}

// Discover the available tools based on the user preferences
func Discover(cfg *config.Config, policies map[string]any) (availableTools []api.ToolsProvider, notAvailableTools []api.ToolsProvider) {
	for _, provider := range providers {
		if provider.IsAvailable(cfg, policies[provider.Attributes().Name()]) {
			availableTools = append(availableTools, provider)
		} else {
			notAvailableTools = append(notAvailableTools, provider)
		}
	}
	slices.SortFunc(availableTools, func(a, b api.ToolsProvider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	slices.SortFunc(notAvailableTools, func(a, b api.ToolsProvider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return availableTools, notAvailableTools
}

func GetDefaultPolicies() map[string]any {
	policies := make(map[string]any)
	for _, provider := range providers {
		providerPolicies := provider.GetDefaultPolicies()
		policies[provider.Attributes().Name()] = providerPolicies
	}
	return policies
}
