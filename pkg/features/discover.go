package features

import (
	"fmt"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Features struct {
	Inferences             []api.InferenceProvider `json:"inferences"`             // List of available inference providers
	InferencesNotAvailable []api.InferenceProvider `json:"inferencesNotAvailable"` // List of not available inference providers
	Inference              *api.InferenceProvider  `json:"inference"`              // The selected inference provider based on user preferences or auto-detection, or nil if no inference provider is selected
	Tools                  []api.ToolsProvider     `json:"tools"`                  // List of available tools
	ToolsNotAvailable      []api.ToolsProvider     `json:"toolsNotAvailable"`      // List of not available tools
}

// ToHumanReadable converts the features to a human-readable string representation.
func (f *Features) ToHumanReadable() string {
	ret := &strings.Builder{}
	_, _ = fmt.Fprint(ret, "Available Inference Providers:\n")
	for _, provider := range f.Inferences {
		_, _ = fmt.Fprint(ret, toHumanReadable(provider))
	}
	_, _ = fmt.Fprint(ret, "Not Available Inference Providers:\n")
	for _, provider := range f.InferencesNotAvailable {
		_, _ = fmt.Fprint(ret, toHumanReadable(provider))
	}
	if f.Inference != nil {
		_, _ = fmt.Fprintf(ret, "Selected Inference Provider: %s\n", (*f.Inference).Attributes().Name())
	}
	_, _ = fmt.Fprint(ret, "Available Tools Providers:\n")
	for _, provider := range f.Tools {
		_, _ = fmt.Fprint(ret, toHumanReadable(provider))
	}
	_, _ = fmt.Fprint(ret, "Not Available Tools Providers:\n")
	for _, provider := range f.ToolsNotAvailable {
		_, _ = fmt.Fprint(ret, toHumanReadable(provider))
	}
	return ret.String()
}

func toHumanReadable[A api.FeatureAttributes](p api.Feature[A]) string {
	ret := &strings.Builder{}
	_, _ = fmt.Fprintf(ret, "  - %s\n", p.Attributes().Name())
	_, _ = fmt.Fprintf(ret, "    Description: %s\n", p.Attributes().Description())
	_, _ = fmt.Fprintf(ret, "    Reason: %s\n", p.Reason())
	return ret.String()
}

func Discover(cfg *config.Config, policies *policies.Policies) *Features {
	availableInferences, notAvailableInferences := inference.Discover(cfg, nil) // TODO: pass preferences for inference

	var selectedInference *api.InferenceProvider
	if cfg.Inference != nil {
		for _, i := range availableInferences {
			if i.Attributes().Name() == *cfg.Inference {
				selectedInference = &i
				break
			}
		}
	} else if len(availableInferences) > 0 {
		// TODO: Implement user preferences or auto-detection logic to select the best inference
		// For now, we just select the first available inference
		selectedInference = &availableInferences[0]
	}
	toolsPolicies := map[string]any{}
	if policies != nil {
		toolsPolicies = policies.Tools
	}
	availableTools, notAvailableTools := tools.Discover(cfg, toolsPolicies)
	return &Features{
		Inferences:             availableInferences,
		InferencesNotAvailable: notAvailableInferences,
		Inference:              selectedInference,
		Tools:                  availableTools,
		ToolsNotAvailable:      notAvailableTools,
	}
}

func GetDefaultPolicies() map[string]any {
	return map[string]any{
		"tools": tools.GetDefaultPolicies(),
	}
}
