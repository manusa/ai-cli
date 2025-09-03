package policies

import "github.com/manusa/ai-cli/pkg/api"

func (p *Provider) IsToolEnabledByPolicies(feature api.Feature[api.ToolsAttributes], policies *api.Policies) (value bool, enforced bool) {
	if policies == nil {
		return false, false
	}
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].Enabled != nil {
		return *policies.Tools.Provider[providerName].Enabled, true
	}

	if policies.Tools.Enabled != nil {
		return *policies.Tools.Enabled, true
	}

	// TODO should be disabled if read-only/... is set by policy and provider does not support it

	return false, false
}
