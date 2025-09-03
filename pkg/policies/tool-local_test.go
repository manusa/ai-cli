package policies

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestIsToolLocalByPolicies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		feature      api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions]
		policiesToml string
		expected     bool
	}{
		{
			name:     "tool not local by default",
			expected: false,
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
			name:     "provider local by name",
			expected: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools.provider.provider1]
local = true
`,
		},
		{
			name:     "provider not local by name",
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
local = true

[tools.provider.provider1]
local = false
`,
		},
		{
			name:     "provider local globally",
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
local = true
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			policies, err := readToml(tt.policiesToml)
			assert.NoError(t, err)
			actual := provider.IsToolLocalByPolicies(tt.feature, policies)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
