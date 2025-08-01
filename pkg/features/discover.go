package features

import (
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Features struct {
	Inferences []inference.Provider `json:"inferences"` // List of available inference providers
	Inference  *inference.Provider  `json:"inference"`  // The selected inference provider based on user preferences or auto-detection, or nil if no inference provider is selected
	Tools      []tools.Provider     `json:"tools"`      // List of tools available from the selected inference provider
}

func Discover(cfg *config.Config) *Features {
	inferences := inference.Discover(cfg)

	var selectedInference *inference.Provider
	if cfg.Inference != nil {
		for _, inference := range inferences {
			if inference.Attributes().Name() == *cfg.Inference {
				selectedInference = &inference
				break
			}
		}
	} else if len(inferences) > 0 {
		// TODO: Implement user preferences or auto-detection logic to select the best inference
		// For now, we just select the first available inference
		selectedInference = &inferences[0]
	}
	return &Features{
		Inferences: inferences,
		Inference:  selectedInference,
		Tools:      tools.Discover(cfg),
	}
}
