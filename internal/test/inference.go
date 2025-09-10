package test

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
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

func WithInferenceLlm(llm model.ToolCallingChatModel) InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.Llm = llm
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

func (i *InferenceProvider) Initialize(_ context.Context) {
	i.Initialized = true
}

func (i *InferenceProvider) GetInference(_ context.Context) (model.ToolCallingChatModel, error) {
	return i.Llm, nil
}
