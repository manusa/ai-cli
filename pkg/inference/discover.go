package inference

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]InferenceProvider{}

type InferenceAttributes struct {
	Name    string
	Distant bool
}

type InferenceProvider interface {
	Attributes() InferenceAttributes
	IsAvailable(cfg *config.Config) bool
	GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error)
}

// Register a new inference provider
func Register(provider InferenceProvider) {
	if _, ok := providers[provider.Attributes().Name]; ok {
		panic(fmt.Sprintf("inference provider already registered: %s", provider.Attributes().Name))
	}
	providers[provider.Attributes().Name] = provider
}

// cleanup for tests
func cleanup() {
	providers = map[string]InferenceProvider{}
}

// Discover the best inference based on the user preferences
func Discover(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	inferences, err := getAvailableInferences(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get available inferences: %w", err)
	}

	if len(inferences) == 0 {
		return nil, fmt.Errorf("no inference found")
	}

	// TODO: select the best inference based on the attributes compared to user preferences
	inferenceAttributes := inferences[0]

	provider, ok := providers[inferenceAttributes.Name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", inferenceAttributes.Name)
	}
	inference, err := provider.GetInference(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get inference: %w", err)
	}

	return inference, nil
}

// getAvailableInferences gets all available inferences from all providers
func getAvailableInferences(cfg *config.Config) ([]InferenceAttributes, error) {
	inferences := []InferenceAttributes{}
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			inferences = append(inferences, provider.Attributes())
		}
	}
	return inferences, nil
}
