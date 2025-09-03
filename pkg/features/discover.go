package features

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
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

// ToJSON converts the features to a generic JSON string representation.
func (f *Features) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(f, "", "  ")
	return string(bytes), err
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

func Discover(ctx context.Context) *Features {
	unregisterDisabledInferences(ctx)
	unregisterDisabledTools(ctx)
	cfg := config.GetConfig(ctx)
	availableInferences, notAvailableInferences := classify(inference.Initialize(ctx)) // TODO: pass preferences for inference

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

	availableTools, notAvailableTools := classify(tools.Initialize(ctx))
	return &Features{
		Inferences:             availableInferences,
		InferencesNotAvailable: notAvailableInferences,
		Inference:              selectedInference,
		Tools:                  availableTools,
		ToolsNotAvailable:      notAvailableTools,
	}
}

func unregisterDisabledInferences(ctx context.Context) {
	ctxPolicies := policies.GetPolicies(ctx)
	for name, provider := range inference.GetProviders() {
		if ctxPolicies != nil && !policies.PoliciesProvider.IsInferenceEnabledByPolicies(provider, ctxPolicies) {
			inference.Unregister(name)
			continue
		}
	}
}

func unregisterDisabledTools(ctx context.Context) {
	ctxPolicies := policies.GetPolicies(ctx)
	for name, provider := range tools.GetProviders() {
		if ctxPolicies != nil && !policies.PoliciesProvider.IsToolEnabledByPolicies(provider, ctxPolicies) {
			tools.Unregister(name)
			continue
		}
	}
}

func classify[A api.FeatureAttributes, F api.Feature[A]](providers []F) (availableFeatures []F, notAvailableFeatures []F) {
	availableFeatures = []F{}
	notAvailableFeatures = []F{}
	for _, provider := range providers {
		if provider.IsAvailable() {
			availableFeatures = append(availableFeatures, provider)
		} else {
			notAvailableFeatures = append(notAvailableFeatures, provider)
		}
	}
	slices.SortFunc(availableFeatures, api.FeatureSorter)
	slices.SortFunc(notAvailableFeatures, api.FeatureSorter)
	return
}
