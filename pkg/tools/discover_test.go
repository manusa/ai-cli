package tools

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/ai-cli/pkg/config"
)

type DiscoverTestSuite struct {
	suite.Suite
}

func (s *DiscoverTestSuite) SetupTest() {
	Clear()
}

func (s *DiscoverTestSuite) TestRegister() {
	// Registering a provider should add it to the providers map
	s.Run("Registering a provider adds it to the providers map", func() {
		Register(&test.ToolsProvider{
			BasicToolsProvider: api.BasicToolsProvider{
				BasicToolsAttributes: api.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "testProvider",
						FeatureDescription: "Test Provider",
					},
				},
				Available: true,
			},
		})
		s.Contains(providers, "testProvider",
			"expected provider %s to be registered in the providers %v", "testProvider", providers)
	})
	// Registering a provider with the same name should panic
	s.Run("Registering a provider with the same name panics", func() {
		provider := &test.ToolsProvider{
			BasicToolsProvider: api.BasicToolsProvider{
				BasicToolsAttributes: api.BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "duplicateProvider",
						FeatureDescription: "Test Provider",
					},
				},
				Available: true,
			},
		}
		Register(provider)
		s.Panics(func() {
			Register(provider)
		}, "expected panic when registering a provider with the same name")
	})
	// Registering a nil provider should panic
	s.Run("Registering a nil provider panics", func() {
		s.Panics(func() {
			Register(nil)
		}, "expected panic when registering a nil provider")
	})
}

func (s *DiscoverTestSuite) TestInitialize() {
	provider := test.NewToolsProvider("the-provider")
	Register(provider)
	Initialize(nil, nil)
	s.Run("Initialize calls Initialize on all providers", func() {
		s.True(provider.Initialized, "expected provider to be initialized")
	})
}

func (s *DiscoverTestSuite) TestMarshalling() {
	Register(test.NewToolsProvider(
		"provider-one",
		test.WithToolsAvailable(),
		func(provider *test.ToolsProvider) {
			provider.FeatureDescription = "Test Provider"
		},
	))
	Register(test.NewToolsProvider(
		"provider-two",
		test.WithToolsAvailable(),
		func(provider *test.ToolsProvider) {
			provider.FeatureDescription = "Test Provider"
		},
	))
	discoveredTools := Initialize(config.New(), nil)
	bytes, err := json.Marshal(discoveredTools)
	s.Run("Marshalling returns no error", func() {
		s.Nil(err, "expected no error when marshalling inferences")
	})
	s.Run("Marshalling returns expected JSON", func() {
		s.JSONEq(`[{"description":"Test Provider","name":"provider-one","reason":""},{"description":"Test Provider","name":"provider-two","reason":""}]`, string(bytes),
			"expected JSON to match the expected format")
	})
}

func (s *DiscoverTestSuite) TestGetDefaultPolicies() {
	Register(test.NewToolsProvider(
		"provider-one",
		func(provider *test.ToolsProvider) {
			provider.Policies = map[string]any{
				"provider-one-policy": "provider-one-policy-value",
			}
		},
	))
	Register(test.NewToolsProvider(
		"provider-two",
		func(provider *test.ToolsProvider) {
			provider.Policies = map[string]any{
				"provider-two-policy": "provider-two-policy-value",
			}
		},
	))
	s.Run("GetDefaultPolicies returns expected policies", func() {
		policies := GetDefaultPolicies()
		fmt.Printf("policies: %+v\n", policies)
		s.Equal(map[string]any{"provider-one-policy": "provider-one-policy-value"}, policies["provider-one"], "expected the provider-one policy to be returned")
		s.Equal(map[string]any{"provider-two-policy": "provider-two-policy-value"}, policies["provider-two"], "expected the provider-two policy to be returned")
	})
}

func TestDiscover(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
