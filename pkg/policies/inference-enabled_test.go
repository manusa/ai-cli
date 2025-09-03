package policies

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestIsInferenceEnabledByPolicies(t *testing.T) {
	for _, tt := range []struct {
		name         string
		feature      api.Feature[api.InferenceAttributes, api.InferenceInitializeOptions]
		policiesToml string
		expected     bool
	}{
		{
			name:     "inference enabled by default",
			expected: true,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: ``,
		},
		{
			name:     "provider disabled by name",
			expected: false,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[inferences.provider.provider1]
enabled = false
`,
		},
		{
			name:     "provider enabled by name",
			expected: true,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[inferences]
enabled = false

[inferences.provider.provider1]
enabled = true
`,
		},
		{
			name:     "remote provider disabled by remote property",
			expected: false,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
						LocalAttr:              false,
					},
				},
			},
			policiesToml: `
[inferences.property.remote]
enabled = false
`,
		},
		{
			name:     "remote provider enabled by remote property",
			expected: true,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
						LocalAttr:              false,
					},
				},
			},
			policiesToml: `
[inferences]
enabled = false

[inferences.property.remote]
enabled = true
`,
		},

		{
			name:     "provider disabled globally",
			expected: false,
			feature: &test.InferenceProvider{
				BasicInferenceProvider: api.BasicInferenceProvider{
					BasicInferenceAttributes: api.BasicInferenceAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider1"},
					},
				},
			},
			policiesToml: `
[inferences]
enabled = false
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{}
			policies, err := readToml(tt.policiesToml)
			assert.NoError(t, err)
			actual := provider.IsInferenceEnabledByPolicies(tt.feature, policies)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
