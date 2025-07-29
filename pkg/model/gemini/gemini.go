package gemini

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	einoModel "github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/model"
	"google.golang.org/genai"
)

type GeminiProvider struct{}

var geminiProvider = GeminiProvider{}

func init() {
	model.Register(geminiProvider)
}

func (geminiProvider GeminiProvider) Attributes() model.ModelAttributes {
	return model.ModelAttributes{
		Name:    "gemini",
		Distant: true,
	}
}

func (geminiProvider GeminiProvider) IsAvailable(cfg *config.Config) bool {
	return cfg.GoogleApiKey != ""
}

func (geminiProvider GeminiProvider) GetModel(ctx context.Context, cfg *config.Config) (einoModel.ToolCallingChatModel, error) {
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
}
