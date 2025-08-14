package features

import (
	"fmt"
	"strings"

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

// ToHumanReadable converts the features to a human-readable string representation.
func (f *Features) ToHumanReadable() string {
	ret := &strings.Builder{}
	_, _ = fmt.Fprint(ret, "Available Inference Providers:\n")
	for _, provider := range f.Inferences {
		_, _ = fmt.Fprint(ret, toHumanReadableInferenceProvider(provider))
	}
	_, _ = fmt.Fprint(ret, "Not Available Inference Providers:\n")
	for _, provider := range f.InferencesNotAvailable {
		_, _ = fmt.Fprint(ret, toHumanReadableInferenceProvider(provider))
	}
	if f.Inference != nil {
		_, _ = fmt.Fprintf(ret, "Selected Inference Provider: %s\n", (*f.Inference).Attributes().Name())
	}
	_, _ = fmt.Fprint(ret, "Available Tools Providers:\n")
	for _, provider := range f.Tools {
		_, _ = fmt.Fprint(ret, toHumanReadableToolsProvider(provider))
	}
	_, _ = fmt.Fprint(ret, "Not Available Tools Providers:\n")
	for _, provider := range f.ToolsNotAvailable {
		_, _ = fmt.Fprint(ret, toHumanReadableToolsProvider(provider))
	}
	return ret.String()
}

func toHumanReadableInferenceProvider(provider inference.Provider) string {
	ret := &strings.Builder{}
	_, _ = fmt.Fprintf(ret, "  - %s\n", provider.Attributes().Name())
	reason := provider.Data().Reason
	if reason != "" {
		_, _ = fmt.Fprintf(ret, "    Reason: %s\n", reason)
	}
	return ret.String()
}

func toHumanReadableToolsProvider(provider tools.Provider) string {
	ret := &strings.Builder{}
	_, _ = fmt.Fprintf(ret, "  - %s\n", provider.Attributes().Name())
	reason := provider.Data().Reason
	if reason != "" {
		_, _ = fmt.Fprintf(ret, "    Reason: %s\n", reason)
	}
	return ret.String()
}

func Discover(cfg *config.Config) *Features {
	availableInferences, notAvailableInferences := inference.Discover(cfg)

	var selectedInference *inference.Provider
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
	availableTools, notAvailableTools := tools.Discover(cfg)
	return &Features{
		Inferences:             availableInferences,
		InferencesNotAvailable: notAvailableInferences,
		Inference:              selectedInference,
		Tools:                  availableTools,
		ToolsNotAvailable:      notAvailableTools,
	}
}
