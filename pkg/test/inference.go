package test

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
)

type InferenceProviderOption func(*InferenceProvider)

func WithInferenceAvailable() InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.Available = true
	}
}

func WithInferenceLocal() InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.LocalAttr = true
	}
}

func WithInferencePublic() InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.PublicAttr = true
	}
}

func NewInferenceProvider(name string, options ...InferenceProviderOption) *InferenceProvider {
	p := &InferenceProvider{
		BasicInferenceProvider: api.BasicInferenceProvider{
			BasicInferenceAttributes: api.BasicInferenceAttributes{
				BasicFeatureAttributes: api.BasicFeatureAttributes{
					FeatureName: name,
				},
			},
		},
	}
	for _, option := range options {
		option(p)
	}
	return p
}

type InferenceProvider struct {
	api.BasicInferenceProvider
	Initialized bool                       `json:"-"`
	Llm         model.ToolCallingChatModel `json:"-"`
}

func (i *InferenceProvider) Initialize(_ *config.Config, _ any) {
	i.Initialized = true
}

func (i *InferenceProvider) GetDefaultPolicies() map[string]any {
	return nil
}

func (i *InferenceProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return i.Llm, nil
}
