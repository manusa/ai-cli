package features

import (
	"context"
	"fmt"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Features struct {
	Inferences             []inference.Provider `json:"inferences"`             // List of available inference providers
	InferencesNotAvailable []inference.Provider `json:"inferencesNotAvailable"` // List of not available inference providers
	Inference              *inference.Provider  `json:"inference"`              // The selected inference provider based on user preferences or auto-detection, or nil if no inference provider is selected
	Tools                  []tools.Provider     `json:"tools"`                  // List of available tools
	ToolsNotAvailable      []tools.Provider     `json:"toolsNotAvailable"`      // List of not available tools
}

func Discover(ctx context.Context, cfg *config.Config) *Features {
	availableInferences, notAvailableInferences := inference.Discover(cfg)

	var selectedInference *inference.Provider
	if cfg.Inference != nil {
		for _, inference := range availableInferences {
			if inference.Attributes().Name() == *cfg.Inference {
				selectedInference = &inference
				break
			}
		}
	} else if len(availableInferences) > 0 {
		// TODO: Implement user preferences or auto-detection logic to select the best inference
		// For now, we just select the first available inference
		selectedInference = &availableInferences[0]
	}
	availableTools, notAvailableTools := tools.Discover(cfg)

	if selectedInference != nil {
		avail, notAvail, err := discoverWithModel(ctx, cfg, *selectedInference, availableTools)
		if err == nil {
			availableTools = append(availableTools, avail...)
			notAvailableTools = append(notAvailableTools, notAvail...)
		}
	}

	return &Features{
		Inferences:             availableInferences,
		InferencesNotAvailable: notAvailableInferences,
		Inference:              selectedInference,
		Tools:                  availableTools,
		ToolsNotAvailable:      notAvailableTools,
	}
}

func discoverWithModel(ctx context.Context, cfg *config.Config, inference inference.Provider, discoveredTools []tools.Provider) (availableTools []tools.Provider, notAvailableTools []tools.Provider, err error) {
	llm, err := inference.GetInference(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inference: %w", err)
	}

	availableTools, notAvailableTools = tools.DiscoverWithModel(ctx, cfg, llm, discoveredTools)
	return
}
