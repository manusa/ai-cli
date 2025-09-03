package policies

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestIsToolEnabledByPolicies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		feature      api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions]
		policiesToml string
		expected     bool
	}{
		{
			name:     "tool enabled by default",
			expected: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: ``,
		},
		{
			name:     "provider disabled by name",
			expected: false,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools.provider.provider1]
enabled = false
`,
		},
		{
			name:     "provider enabled by name",
			expected: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools]
enabled = false

[tools.provider.provider1]
enabled = true
`,
		},
		{
			name:     "provider disabled globally",
			expected: false,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools]
enabled = false
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			policies, err := readToml(tt.policiesToml)
			assert.NoError(t, err)
			actual := provider.IsToolEnabledByPolicies(tt.feature, policies)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
