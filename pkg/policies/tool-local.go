package policies

import "github.com/manusa/ai-cli/pkg/api"

const (
	DefaultToolLocal = false
)

func (p *Provider) IsToolLocalByPolicies(feature api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions], policies *api.Policies) bool {
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].Local != nil {
		return *policies.Tools.Provider[providerName].Local
	}

	if policies.Tools.Local != nil {
		return *policies.Tools.Local
	}

	return DefaultToolLocal
}
