package policies

import "github.com/manusa/ai-cli/pkg/api"

const (
	DefaultToolReadonly = false
)

func (p *Provider) IsToolReadonlyByPolicies(feature api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions], policies *api.Policies) bool {
	providerName := feature.Attributes().Name()
	if policies.Tools.Provider[providerName].ReadOnly != nil {
		return *policies.Tools.Provider[providerName].ReadOnly
	}

	if policies.Tools.ReadOnly != nil {
		return *policies.Tools.ReadOnly
	}

	return DefaultToolReadonly
}
