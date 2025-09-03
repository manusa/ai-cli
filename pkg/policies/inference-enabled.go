package policies

import "github.com/manusa/ai-cli/pkg/api"

func (p *Provider) IsInferenceEnabledByPolicies(feature api.Feature[api.InferenceAttributes], policies *api.Policies) (value bool, enforced bool) {
	if policies == nil {
		return false, false
	}
	providerName := feature.Attributes().Name()
	if policies.Inferences.Provider[providerName].Enabled != nil {
		return *policies.Inferences.Provider[providerName].Enabled, true
	}

	providerLocal := feature.Attributes().Local()
	if policies.Inferences.Property.Remote.Enabled != nil {
		if !*policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return false, true
		}
		if *policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return true, true
		}
	}

	if policies.Inferences.Enabled != nil {
		return *policies.Inferences.Enabled, true
	}

	return false, false
}
