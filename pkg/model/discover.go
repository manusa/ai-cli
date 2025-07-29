package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

var providers = map[string]ModelProvider{}

type ModelAttributes struct {
	Name    string
	Distant bool
}

type ModelProvider interface {
	Attributes() ModelAttributes
	IsAvailable(cfg *config.Config) bool
	GetModel(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error)
}

type Model struct {
	ModelAttributes
	Model model.ToolCallingChatModel
}

// Register a new model provider
func Register(provider ModelProvider) {
	if _, ok := providers[provider.Attributes().Name]; ok {
		panic(fmt.Sprintf("model provider already registered: %s", provider.Attributes().Name))
	}
	providers[provider.Attributes().Name] = provider
}

// cleanup for tests
func cleanup() {
	providers = map[string]ModelProvider{}
}

// Discover the best model based on the user preferences
func Discover(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	models, err := getAvailableModels(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no model found")
	}

	// TODO: select the best model based on the attributes compared to user preferences
	modelAttributes := models[0]

	provider, ok := providers[modelAttributes.Name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", modelAttributes.Name)
	}
	model, err := provider.GetModel(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	return model, nil
}

// getAvailableModels gets all available models from all providers
func getAvailableModels(cfg *config.Config) ([]ModelAttributes, error) {
	models := []ModelAttributes{}
	for _, provider := range providers {
		if provider.IsAvailable(cfg) {
			models = append(models, provider.Attributes())
		}
	}
	return models, nil
}
