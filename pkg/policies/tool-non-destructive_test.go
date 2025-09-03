package policies

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestIsToolNonDestructiveByPolicies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		feature      api.Feature[api.ToolsAttributes]
		policiesToml string
		expected     bool
		enforced     bool
	}{
		{
			name:     "tool not non destructive by default",
			expected: false,
			enforced: false,
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
			name:     "provider non destructive by name",
			expected: true,
			enforced: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools.provider.provider1]
non-destructive = true
`,
		},
		{
			name:     "provider not non destructive by name",
			expected: false,
			enforced: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools]
non-destructive = true

[tools.provider.provider1]
non-destructive = false
`,
		},
		{
			name:     "provider non destructive globally",
			expected: true,
			enforced: true,
			feature: &test.ToolsProvider{
				BasicToolsProvider: api.BasicToolsProvider{
					BasicToolsAttributes: api.BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[tools]
non-destructive = true
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			policies, err := ReadToml(tt.policiesToml)
			assert.NoError(t, err)
			actual, enforced := provider.IsToolNonDestructiveByPolicies(tt.feature, policies)
			assert.Equal(t, tt.enforced, enforced)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
