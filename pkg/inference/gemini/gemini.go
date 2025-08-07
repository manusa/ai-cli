package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"google.golang.org/genai"
)

type Provider struct {
	inference.BasicInferenceProvider
}

var _ inference.Provider = &Provider{}

func (geminiProvider *Provider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "gemini",
		},
		Local:  false,
		Public: true,
	}
}

func (geminiProvider *Provider) Data() inference.Data {
	return inference.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: geminiProvider.Reason,
		},
		Models: geminiProvider.Models,
	}
}

func (geminiProvider *Provider) IsAvailable(cfg *config.Config) bool {
	available := cfg.GoogleApiKey != ""
	if available {
		geminiProvider.Reason = "GEMINI_API_KEY is set"
		geminiProvider.Models = []string{"gemini-2.0-flash"}
	} else {
		geminiProvider.Reason = "GEMINI_API_KEY is not set"
	}
	return available
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

func (geminiProvider *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(inference.Report{
		Attributes: geminiProvider.Attributes(),
		Data:       geminiProvider.Data(),
	})
}

var instance = &Provider{}

func init() {
	inference.Register(instance)
}
