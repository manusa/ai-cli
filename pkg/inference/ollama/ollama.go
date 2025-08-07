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
		Local:  true,
		Public: false,
	}
}

func (ollamaProvider *Provider) Data() inference.Data {
	return inference.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: ollamaProvider.Reason,
		},
		Models: ollamaProvider.Models,
	}
}

func (ollamaProvider *Provider) GetModels(_ context.Context, _ *config.Config) ([]string, error) {
	resp, err := http.Get(ollamaProvider.baseURL() + "/v1/models")
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

func (ollamaProvider *Provider) IsAvailable(cfg *config.Config) bool {
	baseURL := ollamaProvider.baseURL()
	isBaseURLConfigured := ollamaProvider.isBaseURLConfigured()
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		if isBaseURLConfigured {
			ollamaProvider.Reason = fmt.Sprintf("%s defined by the %s environment variable is not accessible", baseURL, ollamaHostEnvVar)
		} else {
			ollamaProvider.Reason = fmt.Sprintf("%s is not accessible", baseURL)
		}
		return false
	}
	_ = resp.Body.Close()
	available := resp.StatusCode == http.StatusOK
	if available {
		if isBaseURLConfigured {
			ollamaProvider.Reason = fmt.Sprintf("ollama is accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
		} else {
			ollamaProvider.Reason = fmt.Sprintf("ollama is accessible at %s", baseURL)
		}
		models, err := ollamaProvider.GetModels(context.Background(), cfg)
		if err == nil {
			ollamaProvider.Models = models
		}
	} else {
		if isBaseURLConfigured {
			ollamaProvider.Reason = fmt.Sprintf("ollama is not accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
		} else {
			ollamaProvider.Reason = fmt.Sprintf("ollama is not accessible at %s", baseURL)
		}
	}
	return available
}

func (ollamaProvider *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	model := defaultModel
	if cfg.Model != nil {
		model = *cfg.Model
	}
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: ollamaProvider.baseURL(),
		Model:   model,
	})
}

func (ollamaProvider *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(inference.Report{
		Attributes: ollamaProvider.Attributes(),
		Data:       ollamaProvider.Data(),
	})
}

func (ollamaProvider *Provider) baseURL() string {
	if baseURL := os.Getenv(ollamaHostEnvVar); baseURL != "" {
		return baseURL
	}
	return defaultBaseURL
}

func (ollamaProvider *Provider) isBaseURLConfigured() bool {
	return os.Getenv(ollamaHostEnvVar) != ""
}

var instance = &Provider{}

func init() {
	inference.Register(instance)
}
