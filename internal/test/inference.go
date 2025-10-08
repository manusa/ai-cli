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

func WithGetModel(getModel func() (string, error)) InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.getModel = getModel
	}
}

func WithSupportsSetup() InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.SupportsSetupAttr = true
	}
}

func WithInstallHelp(installHelp func() error) InferenceProviderOption {
	return func(i *InferenceProvider) {
		i.installHelp = installHelp
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
	getModel    func() (string, error)     `json:"-"`
	installHelp func() error               `json:"-"`
}

func (i *InferenceProvider) Initialize(_ context.Context) {
	i.Initialized = true
}

func (i *InferenceProvider) GetInference(_ context.Context) (model.ToolCallingChatModel, error) {
	return i.Llm, nil
}

func (i *InferenceProvider) InstallHelp() error {
	if i.installHelp == nil {
		return nil
	}
	return i.installHelp()
}

func (i *InferenceProvider) GetModel(_ context.Context) (string, error) {
	if i.getModel == nil {
		return "", nil
	}
	return i.getModel()
}
