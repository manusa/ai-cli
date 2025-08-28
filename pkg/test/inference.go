package test

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

type InferenceProvider struct {
	api.BasicInferenceProvider
	Available bool                       `json:"-"`
	Llm       model.ToolCallingChatModel `json:"-"`
}

func (i *InferenceProvider) IsAvailable(_ *config.Config, _ any) bool {
	return i.Available
}

func (i *InferenceProvider) GetDefaultPolicies() map[string]any {
	return nil
}

func (i *InferenceProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return i.Llm, nil
}
