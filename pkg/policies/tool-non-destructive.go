package policies

import "github.com/manusa/ai-cli/pkg/api"

func (p *Provider) IsToolNonDestructiveByPolicies(feature api.Feature[api.ToolsAttributes], policies *api.Policies) (value bool, enforced bool) {
	if policies == nil {
		return false, false
	}
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].NonDestructive != nil {
		return *policies.Tools.Provider[providerName].NonDestructive, true
	}

	if policies.Tools.NonDestructive != nil {
		return *policies.Tools.NonDestructive, true
	}

	return false, false
}
