package policies

import "github.com/manusa/ai-cli/pkg/api"

const (
	DefaultToolEnabled = true
)

func (p *Provider) IsToolEnabledByPolicies(feature api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions], policies *api.Policies) bool {
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].Enabled != nil {
		return *policies.Tools.Provider[providerName].Enabled
	}

	if policies.Tools.Enabled != nil {
		return *policies.Tools.Enabled
	}

	// TODO should be disabled if read-only/... is set by policy and provider does not support it

	return DefaultToolEnabled
}
