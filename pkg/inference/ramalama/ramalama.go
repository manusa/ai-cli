package ramalama

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
)

type Provider struct {
	api.BasicInferenceProvider
	processes []ramalamaProcess
}

var _ api.InferenceProvider = &Provider{}

// ramalamaProcess is part of the response from the "ramalama ps --format json" command
type ramalamaProcess struct {
	State  string
	Labels map[string]string
}

func (p *Provider) GetModels(_ context.Context) ([]string, error) {
	cmd := exec.Command(p.getRamalamaBinaryName(), "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(output, &p.processes)
	if err != nil {
		return nil, err
	}
	models := make([]string, 0, len(p.processes))
	for _, process := range p.processes {
		models = append(models, process.Labels["ai.ramalama.model"])
	}
	return models, nil
}

func (p *Provider) Initialize(ctx context.Context, _ any) {
	if !config.CommandExists(p.getRamalamaBinaryName()) {
		p.IsAvailableReason = "ramalama is not installed"
		return
	}
	models, err := p.GetModels(ctx)
	if err != nil || len(models) == 0 {
		p.IsAvailableReason = "ramalama is installed but no models are served"
		return
	}
	p.ProviderModels = models

	p.Available = true
	p.IsAvailableReason = "ramalama is serving models"
}

func (p *Provider) GetInference(ctx context.Context) (model.ToolCallingChatModel, error) {
	model := p.ProviderModels[0]
	cfg := config.GetConfig(ctx)
	if cfg.Model != nil {
		model = *cfg.Model
	}
	baseURL, err := p.baseURL(model)
	if err != nil {
		return nil, err
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   model,
	})
}

func (p *Provider) baseURL(model string) (string, error) {
	process := p.getProcessByModel(model)
	if process == nil {
		return "", fmt.Errorf("model %s not found", model)
	}
	url := fmt.Sprintf("http://localhost:%s", process.Labels["ai.ramalama.port"])
	return url, nil
}

func (p *Provider) getProcessByModel(model string) *ramalamaProcess {
	for _, process := range p.processes {
		if process.Labels["ai.ramalama.model"] == model {
			return &process
		}
	}
	return nil
}

func (p *Provider) GetDefaultPolicies() map[string]any {
	return nil
}

func (p *Provider) getRamalamaBinaryName() string {
	if runtime.GOOS == "windows" {
		return "ramalama.exe"
	}
	return "ramalama"
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "ramalama",
				FeatureDescription: "Ramalama local inference provider",
			},
			LocalAttr:  true,
			PublicAttr: false,
		},
	},
	nil,
}

func init() {
	inference.Register(instance)
}
