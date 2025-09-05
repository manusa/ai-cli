package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

const ollamaHostEnvVar = "OLLAMA_HOST"

var (
	// DefaultBaseURL is the default base URL for the Ollama API (Exposed for testing purposes)
	DefaultBaseURL  = "http://localhost:11434"
	preferredModels = []string{
		"llama3.2:3b",
		"granite3.3:latest",
		"mistral:7b",
	}
)

type Provider struct {
	api.BasicInferenceProvider
}

var _ api.InferenceProvider = &Provider{}

// ModelsList is the response from the /v1/models endpoint
type ModelsList struct {
	Data []struct {
		Id string `json:"id"`
	} `json:"data"`
}

func (p *Provider) GetModels(_ context.Context) ([]string, error) {
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

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.InferenceParameters = cfg.InferenceParameters(p.Attributes().Name())
	}

	baseURL := p.baseURL()
	isBaseURLConfigured := p.isBaseURLConfigured()
	resp, err := http.Get(baseURL + "/v1/models")
	defer func(resp *http.Response) {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}(resp)
	if err != nil || resp.StatusCode != http.StatusOK {
		if isBaseURLConfigured {
			p.IsAvailableReason = fmt.Sprintf("ollama is not accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
		} else {
			p.IsAvailableReason = fmt.Sprintf("ollama is not accessible at %s", baseURL)
		}
		return
	}

	p.Available = true
	if isBaseURLConfigured {
		p.IsAvailableReason = fmt.Sprintf("ollama is accessible at %s defined by the %s environment variable", baseURL, ollamaHostEnvVar)
	} else {
		p.IsAvailableReason = fmt.Sprintf("ollama is accessible at %s", baseURL)
	}
	p.ProviderModels, err = p.GetModels(ctx)
	var selectedModel string
	for _, preferredModel := range preferredModels {
		if err != nil && slices.Contains(p.ProviderModels, preferredModel) {
			selectedModel = preferredModel
			break
		}
	}
	if p.Model == nil {
		p.Model = &selectedModel
	}
}

func (p *Provider) GetInference(ctx context.Context) (model.ToolCallingChatModel, error) {
	return ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: p.baseURL(),
		Model:   *p.Model,
	})
}

func (p *Provider) baseURL() string {
	if baseURL := os.Getenv(ollamaHostEnvVar); baseURL != "" {
		if !strings.HasPrefix(baseURL, "http://") {
			baseURL = "http://" + baseURL
		}
		return baseURL
	}
	return DefaultBaseURL
}

func (p *Provider) isBaseURLConfigured() bool {
	return os.Getenv(ollamaHostEnvVar) != ""
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
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
