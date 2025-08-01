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

type Provider struct{}

var _ inference.Provider = &Provider{}

func (geminiProvider *Provider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "gemini",
		},
		Distant: true,
	}
}

func (geminiProvider *Provider) IsAvailable(cfg *config.Config) bool {
	return cfg.GoogleApiKey != ""
}

func (geminiProvider *Provider) GetModels(ctx context.Context, cfg *config.Config) ([]string, error) {
	return []string{"gemini-2.0-flash"}, nil
}

func (geminiProvider *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
}

var instance = &Provider{}

func init() {
	inference.Register(instance)
}
