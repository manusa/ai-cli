package features

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/fs"
	"github.com/manusa/ai-cli/pkg/tools/kubernetes"
	"github.com/stretchr/testify/assert"
)

type testContext struct {
	originalEnv []string
}

func (c *testContext) beforeEach(t *testing.T) {
	t.Helper()
	c.originalEnv = os.Environ()
	os.Clearenv()
	inference.Clear()
	tools.Clear()
}

func (c *testContext) afterEach(t *testing.T) {
	t.Helper()
	os.Clearenv()
	for _, env := range c.originalEnv {
		if key, value, found := strings.Cut(env, "="); found {
			_ = os.Setenv(key, value)
		}
	}
}

func testCase(t *testing.T, test func(c *testContext)) {
	testCaseWithContext(t, &testContext{}, test)
}

func testCaseWithContext(t *testing.T, ctx *testContext, test func(c *testContext)) {
	ctx.beforeEach(t)
	t.Cleanup(func() { ctx.afterEach(t) })
	test(ctx)
}

type TestProvider struct {
	Available bool `json:"-"`
}

func (t *TestProvider) IsAvailable(_ *config.Config, _ any) bool {
	return t.Available
}

func (t *TestProvider) GetDefaultPolicies() map[string]any {
	return nil
}

type TestInferenceProvider struct {
	inference.BasicInferenceProvider
	TestProvider
}

func (t *TestInferenceProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

type TestToolsProvider struct {
	tools.BasicToolsProvider
	TestProvider
}

func (t *TestToolsProvider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return nil, nil
}

func TestDiscoverInference(t *testing.T) {
	// With one available inference provider, it should return that provider
	testCase(t, func(c *testContext) {
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-unavailable", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: false},
		})
		features := Discover(config.New(), nil)
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected an inference to be returned")
		})
		t.Run("With one available provider Inferences has one provider", func(t *testing.T) {
			assert.Len(t, features.Inferences, 1, "expected one inference provider to be returned")
			assert.Equal(t, "provider-available", features.Inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
		t.Run("With one available provider Inference is set to that provider", func(t *testing.T) {
			assert.Equal(t, "provider-available", (*features.Inference).Attributes().Name(),
				"expected the available provider to be returned")
		})
	})
}

func TestDiscoverInferenceConfiguredProvider(t *testing.T) {
	testCase(t, func(c *testContext) {
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-1", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-2", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-3", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		cfg := config.New()
		cfg.Inference = func(s string) *string {
			return &s
		}("provider-2")
		features := Discover(cfg, nil)
		t.Run("Inference is set to configured provider", func(t *testing.T) {
			assert.Equal(t, "provider-2", (*features.Inference).Attributes().Name(),
				"expected the configured provider to be returned")
		})
	})
}

func TestDiscoverInferenceConfiguredProviderUnknown(t *testing.T) {
	testCase(t, func(c *testContext) {
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-1", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-2", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
			},
			TestProvider{Available: true},
		})
		cfg := config.New()
		cfg.Inference = func(s string) *string {
			return &s
		}("unknown-provider")
		features := Discover(cfg, nil)
		t.Run("Inference IS NOT set", func(t *testing.T) {
			assert.Nil(t, features.Inference, "expected nil inference to be returned")
		})
	})
}

func TestDiscoverTools(t *testing.T) {
	testCase(t, func(c *testContext) {
		tools.Register(&TestToolsProvider{
			tools.BasicToolsProvider{
				BasicToolsAttributes: tools.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
				},
			},
			TestProvider{Available: true},
		})
		features := Discover(config.New(), nil)
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected features to be returned")
		})
		t.Run("With one available Tools provider has one provider", func(t *testing.T) {
			assert.Len(t, features.ToolsNotAvailable, 0, "expected no not available tools provider to be returned")
			assert.Len(t, features.Tools, 1, "expected one tool provider to be returned")
			assert.Equal(t, "provider-available", features.Tools[0].Attributes().Name(),
				"expected provider-available provider to be returned")
		})
	})
}

func TestDiscoverToolsWithEnabledPolicies(t *testing.T) {
	testCase(t, func(c *testContext) {
		tools.Register(&TestToolsProvider{
			tools.BasicToolsProvider{
				BasicToolsAttributes: tools.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
				},
			},
			TestProvider{Available: true},
		})
		structuredPolicies := policies.Policies{
			Tools: map[string]any{
				"provider-available": map[string]any{
					"enabled": true,
				},
			},
		}
		features := Discover(config.New(), &structuredPolicies)
		t.Run("Tools is set to available AND enabled in policies providers", func(t *testing.T) {
			assert.Len(t, features.ToolsNotAvailable, 0, "expected no not available tools provider to be returned")
			assert.Len(t, features.Tools, 1, "expected one tool provider to be returned")
			assert.Equal(t, "provider-available", features.Tools[0].Attributes().Name(),
				"expected fs provider to be returned")
		})
	})
}

func TestDiscoverToolsWithDisabledPolicies(t *testing.T) {
	// TODO: See TODOs about policy centralization.
	t.Skip("Disabled until policies are centralized and evaluated in the features discovery")
	testCase(t, func(c *testContext) {
		tools.Register(&TestToolsProvider{
			tools.BasicToolsProvider{
				BasicToolsAttributes: tools.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
				},
			},
			TestProvider{Available: true},
		})
		structuredPolicies := policies.Policies{
			Tools: map[string]any{
				"provider-available": map[string]any{
					"enabled": false,
				},
			},
		}
		features := Discover(config.New(), &structuredPolicies)
		t.Run("ToolsNotAvailable is set to available AND disabled in policies providers", func(t *testing.T) {
			assert.Len(t, features.Tools, 0, "expected no not available tools provider to be returned")
			assert.Len(t, features.ToolsNotAvailable, 1, "expected one tool provider to be returned")
			assert.Equal(t, "provider-available", features.ToolsNotAvailable[0].Attributes().Name(),
				"expected fs provider to be returned")
		})
	})
}

func TestDiscoverMarshal(t *testing.T) {
	testCase(t, func(c *testContext) {
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "inference-provider-available", FeatureDescription: "Test Provider"},
					LocalAttr:              true,
				},
				IsAvailableReason: "conditions met",
				ProviderModels:    []string{"model-1"},
			},
			TestProvider{Available: true},
		})
		inference.Register(&TestInferenceProvider{
			inference.BasicInferenceProvider{
				BasicInferenceAttributes: inference.BasicInferenceAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "inference-provider-unavailable", FeatureDescription: "Test Provider"},
					LocalAttr:              false,
					PublicAttr:             true,
				},
				IsAvailableReason: "conditions NOT met",
			},
			TestProvider{Available: false},
		})
		tools.Register(&TestToolsProvider{
			tools.BasicToolsProvider{
				BasicToolsAttributes: tools.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "tools-provider-available", FeatureDescription: "Test Provider"},
				},
				IsAvailableReason: "tools conditions met",
			},
			TestProvider{Available: true},
		})
		features := Discover(config.New(), nil)
		bytes, err := json.Marshal(features)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `{`+
				`"inference":{"description":"Test Provider","local":true,"models":["model-1"],"name":"inference-provider-available","public":false,"reason":"conditions met"},`+
				`"inferences":[{"description":"Test Provider","local":true,"models":["model-1"],"name":"inference-provider-available","public":false,"reason":"conditions met"}],`+
				`"inferencesNotAvailable":[{"description":"Test Provider","local":false,"models":null,"name":"inference-provider-unavailable","public":true,"reason":"conditions NOT met"}],`+
				`"tools":[{"description":"Test Provider","name":"tools-provider-available","reason":"tools conditions met"}],`+
				`"toolsNotAvailable":null}`,
				string(bytes),
				"expected JSON to match the expected format")
		})
	})
}

func TestGetDefaultPolicies(t *testing.T) {
	// TODO: See TODOs about policy centralization.
	t.Skip("Disabled until policies are centralized and evaluated in the features discovery")
	testCase(t, func(c *testContext) {
		tools.Register(&fs.Provider{})
		tools.Register(&kubernetes.Provider{})
		policies := GetDefaultPolicies()
		fmt.Printf("policies: %+v\n", policies)
		t.Run("GetDefaultPolicies returns expected policies", func(t *testing.T) {
			fsPolicies := policies["tools"].(map[string]any)["fs"]
			assert.Equal(t, map[string]any{
				"enabled":   false,
				"read-only": false,
			}, fsPolicies, "expected the fs policy to be returned")

			kubernetesPolicies := policies["tools"].(map[string]any)["kubernetes"]
			assert.Equal(t, map[string]any{
				"enabled":             false,
				"read-only":           false,
				"disable-destructive": false,
			}, kubernetesPolicies, "expected the kubernetes policy to be returned")
		})
	})
}
