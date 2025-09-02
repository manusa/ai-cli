package policies

import "github.com/manusa/ai-cli/pkg/api"

const (
	DefaultToolNonDestructive = false
)

func (p *Provider) IsToolNonDestructiveByPolicies(feature api.Feature[api.ToolsAttributes], policies *api.Policies) bool {
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].NonDestructive != nil {
		return *policies.Tools.Provider[providerName].NonDestructive
	}

	if policies.Tools.NonDestructive != nil {
		return *policies.Tools.NonDestructive
	}

	return DefaultToolNonDestructive
}
