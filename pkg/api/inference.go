package api

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

type InferenceProvider interface {
	Feature[InferenceAttributes]
	GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error)
	// Models returns the list of supported models by the inference provider
	Models() []string
}

type InferenceAttributes interface {
	FeatureAttributes
	// Local indicates if the inference provider is a local service
	Local() bool
	// Public indicates if the inference provider is public (e.g. OpenAI, Gemini) or private (e.g. Enterprise internal)
	Public() bool
}

type BasicInferenceProvider struct {
	InferenceProvider `json:"-"`
	BasicInferenceAttributes
	IsAvailableReason string   `json:"reason"`
	ProviderModels    []string `json:"models"`
}

func (p *BasicInferenceProvider) Attributes() InferenceAttributes {
	return &p.BasicInferenceAttributes
}

func (p *BasicInferenceProvider) Reason() string {
	return p.IsAvailableReason
}

func (p *BasicInferenceProvider) Models() []string {
	return p.ProviderModels
}

type BasicInferenceAttributes struct {
	BasicFeatureAttributes
	LocalAttr  bool `json:"local"`
	PublicAttr bool `json:"public"`
}

func (a *BasicInferenceAttributes) Local() bool {
	return a.LocalAttr
}

func (a *BasicInferenceAttributes) Public() bool {
	return a.PublicAttr
}
