package feature

import (
	"fmt"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Features struct {
	Inferences []inference.Provider // List of available inference providers
	Inference  inference.Provider   // The selected inference provider based on user preferences or auto-detection
	Tools      []tools.Provider     // List of tools available from the selected inference provider
}

func Discover(cfg *config.Config) (*Features, error) {
	inferences := inference.Discover(cfg)
	if len(inferences) == 0 {
		return nil, fmt.Errorf("no suitable inference found")
	}
	// TODO: Implement user preferences or auto-detection logic to select the best inference
	// For now, we just select the first available inference
	selectedInference := inferences[0]
	return &Features{
		Inferences: inferences,
		Inference:  selectedInference,
		Tools:      tools.Discover(cfg),
	}, nil
}
