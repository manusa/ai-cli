package policies

import "github.com/manusa/ai-cli/pkg/api"

func (p *Provider) IsToolLocalByPolicies(feature api.Feature[api.ToolsAttributes], policies *api.Policies) (value bool, enforced bool) {
	if policies == nil {
		return false, false
	}
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].Local != nil {
		return *policies.Tools.Provider[providerName].Local, true
	}

	if policies.Tools.Local != nil {
		return *policies.Tools.Local, true
	}

	return false, false
}
