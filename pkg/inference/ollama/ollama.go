package ollama

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

const baseURL = "http://localhost:11434" // TODO: make this configurable
const defaultModel = "llama3.2:3b"

type Provider struct{}

var _ inference.Provider = &Provider{}

// ModelsList is the response from the /v1/models endpoint
type ModelsList struct {
	Data []struct {
		Id string `json:"id"`
	} `json:"data"`
}

func (ollamaProvider *Provider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "ollama",
		},
		Distant: false,
	}
}

func (ollamaProvider *Provider) GetModels(_ context.Context, _ *config.Config) ([]string, error) {
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	modelsList := ModelsList{}
	if err = json.Unmarshal(body, &modelsList); err != nil {
		return nil, err
	}
	modelsNames := make([]string, len(modelsList.Data))
	for i, m := range modelsList.Data {
		modelsNames[i] = m.Id
	}
	return modelsNames, nil
}

func (ollamaProvider *Provider) IsAvailable(_ *config.Config) bool {
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (ollamaProvider *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	model := defaultModel
	if cfg.Model != nil {
		model = *cfg.Model
	}
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: baseURL,
		Model:   model,
	})
}

func (ollamaProvider *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(ollamaProvider.Attributes())
}

var instance = &Provider{}

func init() {
	inference.Register(instance)
}
