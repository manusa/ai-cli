package lmstudio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

const defaultBaseURL = "http://localhost:1234"

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
	resp, err := http.Get(baseURL + "/v1/models")
	if err != nil {
		p.IsAvailableReason = fmt.Sprintf("%s is not accessible", baseURL)
		return false
	}
	_ = resp.Body.Close()
	available := resp.StatusCode == http.StatusOK
	if available {
		p.IsAvailableReason = fmt.Sprintf("LM Studio is accessible at %s", baseURL)
		models, err := p.GetModels(context.Background(), cfg)
		if err == nil {
			p.ProviderModels = models
		}
	} else {
		p.IsAvailableReason = fmt.Sprintf("LM Studio is not accessible at %s", baseURL)
	}
	return available
}

func (p *Provider) GetInference(ctx context.Context, cfg *config.Config) (model.ToolCallingChatModel, error) {
	model := p.ProviderModels[0]
	if cfg.Model != nil {
		model = *cfg.Model
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: fmt.Sprintf("%s/v1", p.baseURL()),
		Model:   model,
	})
}

func (p *Provider) baseURL() string {
	return defaultBaseURL
}

func (p *Provider) GetDefaultPolicies() map[string]any {
	return nil
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "lmstudio",
				FeatureDescription: "LM Studio local inference provider",
			},
			LocalAttr:  true,
			PublicAttr: false,
		},
	},
}

func init() {
	inference.Register(instance)
}
