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

// Default base URL for LM Studio (var can be overridden in tests)
var defaultBaseURL = "http://localhost:1234"

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
	resp, err := http.Get(baseURL + "/v1/models")
	defer func(resp *http.Response) {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}(resp)
	if err != nil {
		p.IsAvailableReason = fmt.Sprintf("LM Studio is not accessible at %s", baseURL)
		return
	}
	if resp.StatusCode != http.StatusOK {
		p.IsAvailableReason = fmt.Sprintf("The server at %s is accessible but is not LM Studio", baseURL)
		return
	}

	p.Available = true
	p.IsAvailableReason = fmt.Sprintf("LM Studio is accessible at %s", baseURL)
	p.ProviderModels, _ = p.GetModels(ctx)
	if p.Model == nil && p.ProviderModels != nil && len(p.ProviderModels) > 0 {
		p.Model = &p.ProviderModels[0]
	}
}

func (p *Provider) GetInference(ctx context.Context) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: fmt.Sprintf("%s/v1", p.baseURL()),
		Model:   *p.Model,
	})
}

func (p *Provider) baseURL() string {
	return defaultBaseURL
}

func (p *Provider) InstallHelp() error {
	return nil
}

var instance = &Provider{
	BasicInferenceProvider: api.BasicInferenceProvider{
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
