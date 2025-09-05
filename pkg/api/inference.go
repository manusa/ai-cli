package api

import (
	"context"

	"github.com/cloudwego/eino/components/model"
)

type InferenceProvider interface {
	Feature[InferenceAttributes]
	GetInference(ctx context.Context) (model.ToolCallingChatModel, error)
	// Models returns the list of supported models by the inference provider
	Models() []string
	SystemPrompt() string
}

type InferenceAttributes interface {
	FeatureAttributes
	// Local indicates if the inference provider is a local service
	Local() bool
	// Public indicates if the inference provider is public (e.g. OpenAI, Gemini) or private (e.g. Enterprise internal)
	Public() bool
}

type InferenceParameters struct {
	Enabled *bool   `json:"-" toml:"enabled"`
	Model   *string `json:"-" toml:"enabled"` // A model to use, if not set, the best model will be used
}

type BasicInferenceProvider struct {
	InferenceProvider `json:"-"`
	BasicInferenceAttributes
	Available         bool     `json:"-"`
	IsAvailableReason string   `json:"reason"`
	ProviderModels    []string `json:"models"`
	InferenceParameters
}

func (p *BasicInferenceProvider) Attributes() InferenceAttributes {
	return &p.BasicInferenceAttributes
}

func (p *BasicInferenceProvider) IsAvailable() bool {
	return p.Available
}

func (p *BasicInferenceProvider) Reason() string {
	return p.IsAvailableReason
}

func (p *BasicInferenceProvider) Models() []string {
	return p.ProviderModels
}

func (p *BasicInferenceProvider) SystemPrompt() string {
	return ""
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
