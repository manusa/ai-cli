package policies

import "github.com/manusa/ai-cli/pkg/api"

const (
	DefaultInferenceEnabled = true
)

func (p *Provider) IsInferenceEnabledByPolicies(feature api.Feature[api.InferenceAttributes, api.InferenceInitializeOptions], policies *api.Policies) bool {
	providerName := feature.Attributes().Name()
	if policies.Inferences.Provider[providerName].Enabled != nil {
		return *policies.Inferences.Provider[providerName].Enabled
	}

	providerLocal := feature.Attributes().Local()
	if policies.Inferences.Property.Remote.Enabled != nil {
		if !*policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return false
		}
		if *policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return true
		}
	}

	if policies.Inferences.Enabled != nil {
		return *policies.Inferences.Enabled
	}

	return DefaultInferenceEnabled
}
