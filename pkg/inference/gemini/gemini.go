package gemini

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"google.golang.org/genai"
)

type Provider struct {
	api.BasicInferenceProvider
}

var _ api.InferenceProvider = &Provider{}

func (p *Provider) Initialize(cfg *config.Config, _ any) {
	p.Available = cfg.GoogleApiKey != ""
	if p.Available {
		p.IsAvailableReason = "GEMINI_API_KEY is set"
		p.ProviderModels = []string{"gemini-2.0-flash"}
	} else {
		p.IsAvailableReason = "GEMINI_API_KEY is not set"
	}
}

func (p *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
}

func (p *Provider) GetDefaultPolicies() map[string]any {
	return nil
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "gemini",
				FeatureDescription: "Google Gemini inference provider",
			},
			LocalAttr:  false,
			PublicAttr: true,
		},
	},
}

func init() {
	inference.Register(instance)
}
