package inference

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]Provider{}

type Attributes struct {
	api.BasicFeatureAttributes
	// TODO: maybe rename to local or remote
	Distant bool
}

type Provider interface {
	api.Feature[Attributes]
	GetModels(ctx context.Context, cfg *config.Config) ([]string, error)
	GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error)
}

// Register a new inference provider
func Register(provider Provider) {
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
	providers = map[string]Provider{}
}

// Discover the available inference providers based on the user preferences
func Discover(cfg *config.Config) []Provider {
	var inferences []Provider
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			inferences = append(inferences, provider)
		}
	}
	return inferences
}
