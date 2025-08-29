package features

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/fs"
	"github.com/manusa/ai-cli/pkg/tools/kubernetes"
	"github.com/stretchr/testify/suite"
)

type TestProvider struct {
}

func (t *TestProvider) Initialize(_ *config.Config, _ any) {}

func (t *TestProvider) GetDefaultPolicies() map[string]any {
	return nil
}

type TestToolsProvider struct {
	api.BasicToolsProvider
	TestProvider
}

func (t *TestToolsProvider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return nil, nil
}

type DiscoverTestSuite struct {
	suite.Suite
	originalEnv []string
}

func (s *DiscoverTestSuite) SetupTest() {
	s.originalEnv = os.Environ()
	os.Clearenv()
	inference.Clear()
	tools.Clear()
}

func (s *DiscoverTestSuite) TearDownTest() {
	os.Clearenv()
	for _, env := range s.originalEnv {
		if key, value, found := strings.Cut(env, "="); found {
			_ = os.Setenv(key, value)
		}
	}
}

func (s *DiscoverTestSuite) TestDiscoverInferenceWithNoProviders() {
	features := Discover(config.New(), nil)
	s.Run("With no providers registered returns empty", func() {
		s.Empty(features.Inferences, "expected no discovered inferences to be returned when no providers are registered")
		s.Empty(features.InferencesNotAvailable, "expected no discovered inferences to be returned when no providers are registered")
		s.Nil(features.Inference, "expected no discovered inference to be returned when no providers are registered")
	})
}

func (s *DiscoverTestSuite) TestDiscoverInferenceWithOneProviderAvailable() {
	// With one available inference provider, it should return that provider
	inference.Register(test.NewInferenceProvider(
		"provider-available",
		test.WithInferenceAvailable(),
		test.WithInferenceLocal(),
	))
	inference.Register(test.NewInferenceProvider(
		"provider-unavailable",
		test.WithInferenceLocal(),
	))
	features := Discover(config.New(), nil)
	s.Run("With one available provider returns features", func() {
		s.NotNil(features, "expected an inference to be returned")
	})
	s.Run("With one available provider Inferences has one provider", func() {
		s.Len(features.Inferences, 1, "expected one inference provider to be returned")
		s.Equal("provider-available", features.Inferences[0].Attributes().Name(),
			"expected the available provider to be returned")
	})
	s.Run("With one available provider InferencesNotAvailable has one unavailable provider", func() {
		s.Len(features.InferencesNotAvailable, 1, "expected one inference provider to be returned")
		s.Equal("provider-unavailable", features.InferencesNotAvailable[0].Attributes().Name(),
			"expected the unavailable provider to be returned")
	})
	s.Run("With one available provider Inference is set to that provider", func() {
		s.Equal("provider-available", (*features.Inference).Attributes().Name(),
			"expected the available provider to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverInferenceConfiguredProvider() {
	inference.Register(test.NewInferenceProvider(
		"provider-1",
		test.WithInferenceAvailable(),
		test.WithInferenceLocal(),
	))
	inference.Register(test.NewInferenceProvider(
		"provider-2",
		test.WithInferenceAvailable(),
	))
	inference.Register(test.NewInferenceProvider(
		"provider-3",
		test.WithInferenceAvailable(),
	))
	cfg := config.New()
	cfg.Inference = func(s string) *string {
		return &s
	}("provider-2")
	features := Discover(cfg, nil)
	s.Run("Inference is set to configured provider", func() {
		s.Equal("provider-2", (*features.Inference).Attributes().Name(),
			"expected the configured provider to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverInferenceConfiguredProviderUnknown() {
	inference.Register(test.NewInferenceProvider(
		"provider-1",
		test.WithInferenceAvailable(),
	))
	inference.Register(test.NewInferenceProvider(
		"provider-2",
		test.WithInferenceAvailable(),
	))
	cfg := config.New()
	cfg.Inference = func(s string) *string {
		return &s
	}("unknown-provider")
	features := Discover(cfg, nil)
	s.Run("Inference IS NOT set", func() {
		s.Nil(features.Inference, "expected nil inference to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverToolsWithNoProviders() {
	features := Discover(config.New(), nil)
	s.Run("With no providers registered returns empty", func() {
		s.Empty(features.Tools, "expected no discovered tools to be returned when no providers are registered")
		s.Empty(features.ToolsNotAvailable, "expected no discovered tools to be returned when no providers are registered")
	})
}

func (s *DiscoverTestSuite) TestDiscoverToolsWithOneProviderAvailable() {
	tools.Register(test.NewToolsProvider("provider-available", test.WithToolsAvailable()))
	tools.Register(test.NewToolsProvider("provider-unavailable"))
	features := Discover(config.New(), nil)
	s.Run("With one available provider returns features", func() {
		s.NotNil(features, "expected features to be returned")
	})
	s.Run("With one available Tools provider has one provider", func() {
		s.Len(features.Tools, 1, "expected one tool provider to be returned")
		s.Equal("provider-available", features.Tools[0].Attributes().Name(),
			"expected provider-available provider to be returned")
	})
	s.Run("With one available provider ToolsNotAvailable has one unavailable provider", func() {
		s.Len(features.ToolsNotAvailable, 1, "expected one tools provider to be returned")
		s.Equal("provider-unavailable", features.ToolsNotAvailable[0].Attributes().Name(),
			"expected the unavailable provider to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverToolsWithEnabledPolicies() {
	tools.Register(test.NewToolsProvider("provider-available", test.WithToolsAvailable()))
	structuredPolicies := policies.Policies{
		Tools: map[string]any{
			"provider-available": map[string]any{
				"enabled": true,
			},
		},
	}
	features := Discover(config.New(), &structuredPolicies)
	s.Run("Tools is set to available AND enabled in policies providers", func() {
		s.Len(features.ToolsNotAvailable, 0, "expected no not available tools provider to be returned")
		s.Len(features.Tools, 1, "expected one tool provider to be returned")
		s.Equal("provider-available", features.Tools[0].Attributes().Name(),
			"expected fs provider to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverToolsWithDisabledPolicies() {
	// TODO: See TODOs about policy centralization.
	s.T().Skip("Disabled until policies are centralized and evaluated in the features discovery")
	tools.Register(&TestToolsProvider{
		BasicToolsProvider: api.BasicToolsProvider{
			BasicToolsAttributes: api.BasicToolsAttributes{
				BasicFeatureAttributes: api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
			},
			Available: true,
		},
	})
	structuredPolicies := policies.Policies{
		Tools: map[string]any{
			"provider-available": map[string]any{
				"enabled": false,
			},
		},
	}
	features := Discover(config.New(), &structuredPolicies)
	s.Run("ToolsNotAvailable is set to available AND disabled in policies providers", func() {
		s.Len(features.Tools, 0, "expected no not available tools provider to be returned")
		s.Len(features.ToolsNotAvailable, 1, "expected one tool provider to be returned")
		s.Equal("provider-available", features.ToolsNotAvailable[0].Attributes().Name(),
			"expected fs provider to be returned")
	})
}

func (s *DiscoverTestSuite) TestDiscoverToJSON() {
	inference.Register(test.NewInferenceProvider(
		"inference-provider-available",
		test.WithInferenceAvailable(),
		test.WithInferenceLocal(),
		func(provider *test.InferenceProvider) {
			provider.FeatureDescription = "Test Provider"
			provider.IsAvailableReason = "conditions met"
			provider.ProviderModels = []string{"model-1"}
		},
	))
	inference.Register(test.NewInferenceProvider(
		"inference-provider-unavailable",
		test.WithInferencePublic(),
		func(provider *test.InferenceProvider) {
			provider.FeatureDescription = "Test Provider"
			provider.IsAvailableReason = "conditions NOT met"
		},
	))
	tools.Register(test.NewToolsProvider(
		"tools-provider-available",
		test.WithToolsAvailable(),
		func(provider *test.ToolsProvider) {
			provider.FeatureDescription = "Test Provider"
			provider.IsAvailableReason = "tools conditions met"
		},
	))
	features := Discover(config.New(), nil)
	jsonString, err := features.ToJSON()
	s.Run("Marshalling returns no error", func() {
		s.Nil(err, "expected no error when marshalling inferences")
	})
	s.Run("Marshalling returns expected JSON", func() {
		s.JSONEq(`{`+
			`"inference":{"description":"Test Provider","local":true,"models":["model-1"],"name":"inference-provider-available","public":false,"reason":"conditions met"},`+
			`"inferences":[{"description":"Test Provider","local":true,"models":["model-1"],"name":"inference-provider-available","public":false,"reason":"conditions met"}],`+
			`"inferencesNotAvailable":[{"description":"Test Provider","local":false,"models":null,"name":"inference-provider-unavailable","public":true,"reason":"conditions NOT met"}],`+
			`"tools":[{"description":"Test Provider","name":"tools-provider-available","reason":"tools conditions met"}],`+
			`"toolsNotAvailable":[]}`,
			jsonString,
			"expected JSON to match the expected format")
	})
}

func (s *DiscoverTestSuite) TestGetDefaultPolicies() {
	// TODO: See TODOs about policy centralization.
	s.T().Skip("Disabled until policies are centralized and evaluated in the features discovery")
	tools.Register(&fs.Provider{})
	tools.Register(&kubernetes.Provider{})
	policies := GetDefaultPolicies()
	fmt.Printf("policies: %+v\n", policies)
	s.Run("GetDefaultPolicies returns expected policies", func() {
		fsPolicies := policies["tools"].(map[string]any)["fs"]
		s.Equal(map[string]any{
			"enabled":   false,
			"read-only": false,
		}, fsPolicies, "expected the fs policy to be returned")

		kubernetesPolicies := policies["tools"].(map[string]any)["kubernetes"]
		s.Equal(map[string]any{
			"enabled":             false,
			"read-only":           false,
			"disable-destructive": false,
		}, kubernetesPolicies, "expected the kubernetes policy to be returned")
	})
}

func TestDiscover(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
