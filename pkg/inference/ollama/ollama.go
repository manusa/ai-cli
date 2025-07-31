package ollama

import (
	"context"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

type OllamaProvider struct{}

var ollamaProvider = OllamaProvider{}

func init() {
	inference.Register(ollamaProvider)
}

func (ollamaProvider OllamaProvider) Attributes() inference.InferenceAttributes {
	return inference.InferenceAttributes{
		Name:    "ollama",
		Distant: false,
	}
}

const baseURL = "http://localhost:11434"

func (ollamaProvider OllamaProvider) IsAvailable(cfg *config.Config) bool {
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (ollamaProvider OllamaProvider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   "llama3.2:3b",
	})
}
