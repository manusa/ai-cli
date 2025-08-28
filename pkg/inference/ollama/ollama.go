package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

const defaultBaseURL = "http://localhost:11434"
const defaultModel = "llama3.2:3b"
const ollamaHostEnvVar = "OLLAMA_HOST"

type Provider struct {
	inference.BasicInferenceProvider
}

var _ api.InferenceProvider = &Provider{}

// ModelsList is the response from the /v1/models endpoint
type ModelsList struct {
	Data []struct {
		Id string `json:"id"`
	} `json:"data"`
}

func (p *Provider) GetModels(_ context.Context, _ *config.Config) ([]string, error) {
	resp, err := http.Get(p.baseURL() + "/v1/models")
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

func (p *Provider) IsAvailable(cfg *config.Config, policies any) bool {
	baseURL := p.baseURL()
	isBaseURLConfigured := p.isBaseURLConfigured()
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		if isBaseURLConfigured {
			p.IsAvailableReason = fmt.Sprintf("%s defined by the %s environment variable is not accessible", baseURL, ollamaHostEnvVar)
		} else {
			p.IsAvailableReason = fmt.Sprintf("%s is not accessible", baseURL)
		}
		return false
	}
	_ = resp.Body.Close()
	available := resp.StatusCode == http.StatusOK
	if available {
		if isBaseURLConfigured {
			p.IsAvailableReason = fmt.Sprintf("ollama is accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
		} else {
			p.IsAvailableReason = fmt.Sprintf("ollama is accessible at %s", baseURL)
		}
		models, err := p.GetModels(context.Background(), cfg)
		if err == nil {
			p.ProviderModels = models
		}
	} else {
		if isBaseURLConfigured {
			p.IsAvailableReason = fmt.Sprintf("ollama is not accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
		} else {
			p.IsAvailableReason = fmt.Sprintf("ollama is not accessible at %s", baseURL)
		}
	}
	return available
}

func (p *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	model := defaultModel
	if cfg.Model != nil {
		model = *cfg.Model
	}
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: p.baseURL(),
		Model:   model,
	})
}

func (p *Provider) baseURL() string {
	if baseURL := os.Getenv(ollamaHostEnvVar); baseURL != "" {
		return baseURL
	}
	return defaultBaseURL
}

func (p *Provider) isBaseURLConfigured() bool {
	return os.Getenv(ollamaHostEnvVar) != ""
}

func (p *Provider) GetDefaultPolicies() map[string]any {
	return nil
}

var instance = &Provider{
	inference.BasicInferenceProvider{
		BasicInferenceAttributes: inference.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "ollama",
				FeatureDescription: "Ollama local inference provider",
			},
			LocalAttr:  true,
			PublicAttr: false,
		},
	},
}

func init() {
	inference.Register(instance)
}
