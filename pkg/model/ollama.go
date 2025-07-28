package model

import (
	"context"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

const baseURL = "http://localhost:11434"

func isOllamaAvailable() bool {
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func GetOllama(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   "llama3.2:3b",
	})
}
