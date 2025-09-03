package policies

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestIsToolReadonlyByPolicies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		feature      api.Feature[api.ToolsAttributes, api.ToolsInitializeOptions]
		policiesToml string
		expected     bool
	}{
		{
			name:     "tool not read-only by default",
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
			name:     "provider read-only by name",
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
read-only = true
`,
		},
		{
			name:     "provider not read-only by name",
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
read-only = true

[tools.provider.provider1]
read-only = false
`,
		},
		{
			name:     "provider read only globally",
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
read-only = true
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			policies, err := readToml(tt.policiesToml)
			assert.NoError(t, err)
			actual := provider.IsToolReadonlyByPolicies(tt.feature, policies)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
