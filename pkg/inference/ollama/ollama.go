package ollama

import (
	"context"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

const baseURL = "http://localhost:11434" // TODO: make this configurable

type Provider struct{}

var _ inference.Provider = &Provider{}

func (ollamaProvider *Provider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "ollama",
		},
		Distant: false,
	}
}

func (ollamaProvider *Provider) IsAvailable(cfg *config.Config) bool {
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (ollamaProvider *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   "llama3.2:3b",
	})
}

var instance = &Provider{}

func init() {
	inference.Register(instance)
}
