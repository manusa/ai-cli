package feature

import (
	"fmt"

	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]FeatureProvider{}

type FeatureAttributes struct {
	Name string
}

type FeatureProvider interface {
	Attributes() FeatureAttributes
	IsAvailable(cfg *config.Config) bool
}

type Feature struct {
	FeatureAttributes
}

// Register a new feature provider
func Register(provider FeatureProvider) {
	if _, ok := providers[provider.Attributes().Name]; ok {
		panic(fmt.Sprintf("feature provider already registered: %s", provider.Attributes().Name))
	}
	providers[provider.Attributes().Name] = provider
}

// cleanup for tests
//func cleanup() {
//	providers = map[string]FeatureProvider{}
//}

// getAvailableModels gets all available models from all providers
func GetAvailableFeatures(cfg *config.Config) ([]FeatureAttributes, error) {
	features := []FeatureAttributes{}
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			features = append(features, provider.Attributes())
		}
	}
	return features, nil
}
