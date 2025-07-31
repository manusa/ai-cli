package gemini

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"google.golang.org/genai"
)

type GeminiProvider struct{}

var geminiProvider = GeminiProvider{}

func init() {
	inference.Register(geminiProvider)
}

func (geminiProvider GeminiProvider) Attributes() inference.InferenceAttributes {
	return inference.InferenceAttributes{
		Name:    "gemini",
		Distant: true,
	}
}

func (geminiProvider GeminiProvider) IsAvailable(cfg *config.Config) bool {
	return cfg.GoogleApiKey != ""
}

func (geminiProvider GeminiProvider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
}
