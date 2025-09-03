package policies

import "github.com/manusa/ai-cli/pkg/api"

func (p *Provider) IsToolReadonlyByPolicies(feature api.Feature[api.ToolsAttributes], policies *api.Policies) (value bool, enforced bool) {
	if policies == nil {
		return false, false
	}
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].ReadOnly != nil {
		return *policies.Tools.Provider[providerName].ReadOnly, true
	}

	if policies.Tools.ReadOnly != nil {
		return *policies.Tools.ReadOnly, true
	}

	return false, false
}
