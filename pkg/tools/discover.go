package tools

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]Provider{}

type Attributes struct {
	api.BasicFeatureAttributes
}

type Provider interface {
	api.Feature[Attributes]
	GetTools(ctx context.Context, cfg *config.Config) ([]*api.Tool, error)
	MarshalJSON() ([]byte, error)
}

// Register a new tools provider
func Register(provider Provider) {
	if _, ok := providers[provider.Attributes().Name()]; ok {
		panic(fmt.Sprintf("tool provider already registered: %s", provider.Attributes().Name()))
	}
	providers[provider.Attributes().Name()] = provider
}

// Clear the registered tools providers (Exposed for testing purposes)
func Clear() {
	providers = map[string]Provider{}
}

// Discover the available tools based on the user preferences
func Discover(cfg *config.Config) []Provider {
	var tools []Provider
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			tools = append(tools, provider)
		}
	}
	slices.SortFunc(tools, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return tools
}
