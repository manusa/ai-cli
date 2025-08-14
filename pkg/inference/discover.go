package inference

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]Provider{}

type BasicInferenceProvider struct {
	api.BasicFeatureProvider
	Models []string `json:"models"` // List of models supported by the inference provider
}

type Attributes struct {
	api.BasicFeatureAttributes
	Local  bool `json:"local"`  // Indicates if the inference provider is a local service
	Public bool `json:"public"` // Indicates if the inference provider is public (e.g. OpenAI, Gemini) or private (e.g. Enterprise internal)
}

type Data struct {
	api.BasicFeatureData
	Models []string `json:"models"`
}

type Report struct {
	Attributes
	Data
}

type Provider interface {
	api.Feature[Attributes, Data]
	GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error)
	MarshalJSON() ([]byte, error)
}

type BasicProvider struct {
	Attributes Attributes `json:"attributes"`
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

// Discover the available and not available inference providers based on the user preferences
func Discover(cfg *config.Config) (availableInferences []Provider, notAvailableInferences []Provider) {
	availableInferences, notAvailableInferences = []Provider{}, []Provider{}
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			availableInferences = append(availableInferences, provider)
		} else {
			notAvailableInferences = append(notAvailableInferences, provider)
		}
	}
	slices.SortFunc(availableInferences, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	slices.SortFunc(notAvailableInferences, func(a, b Provider) int {
		return strings.Compare(a.Attributes().Name(), b.Attributes().Name())
	})
	return availableInferences, notAvailableInferences
}
