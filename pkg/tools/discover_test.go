package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/stretchr/testify/assert"

	"github.com/manusa/ai-cli/pkg/config"
)

type TestProvider struct {
	BasicToolsProvider
	Available bool           `json:"-"`
	Tools     []*api.Tool    `json:"-"`
	Policies  map[string]any `json:"-"`
}

func (t *TestProvider) IsAvailable(_ *config.Config, _ any) bool {
	return t.Available
}

func (t *TestProvider) GetDefaultPolicies() map[string]any {
	return t.Policies
}

func (t *TestProvider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return t.Tools, nil
}

type testContext struct {
}

func (c *testContext) beforeEach(t *testing.T) {
	t.Helper()
	Clear()
}

func (c *testContext) afterEach(t *testing.T) {
	t.Helper()
}

func testCase(t *testing.T, test func(c *testContext)) {
	testCaseWithContext(t, &testContext{}, test)
}

func testCaseWithContext(t *testing.T, ctx *testContext, test func(c *testContext)) {
	ctx.beforeEach(t)
	t.Cleanup(func() { ctx.afterEach(t) })
	test(ctx)
}

func TestRegister(t *testing.T) {
	// Registering a provider should add it to the providers map
	testCase(t, func(c *testContext) {
		t.Run("Registering a provider adds it to the providers map", func(t *testing.T) {
			Register(&TestProvider{
				BasicToolsProvider: BasicToolsProvider{
					BasicToolsAttributes: BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{
							FeatureName:        "testProvider",
							FeatureDescription: "Test Provider",
						},
					},
				},
				Available: true,
			})
			assert.Contains(t, providers, "testProvider",
				"expected provider %s to be registered in the providers %v", "testProvider", providers)
		})
	})
	// Registering a provider with the same name should panic
	testCase(t, func(c *testContext) {
		t.Run("Registering a provider with the same name panics", func(t *testing.T) {
			provider := &TestProvider{
				BasicToolsProvider: BasicToolsProvider{
					BasicToolsAttributes: BasicToolsAttributes{
						BasicFeatureAttributes: api.BasicFeatureAttributes{
							FeatureName:        "duplicateProvider",
							FeatureDescription: "Test Provider",
						},
					},
				},
				Available: true,
			}
			Register(provider)
			assert.Panics(t, func() {
				Register(provider)
			}, "expected panic when registering a provider with the same name")
		})
	})
}

func TestDiscover(t *testing.T) {
	// With no providers registered, it should returns empty
	testCase(t, func(c *testContext) {
		t.Run("With no providers registered returns empty", func(t *testing.T) {
			availableTools, notAvailableTools := Discover(config.New(), nil)
			assert.Empty(t, availableTools, "expected no available tools to be returned when no providers are registered")
			assert.Empty(t, notAvailableTools, "expected no not available tools to be returned when no providers are registered")
		})
	})
	// With one available provider, it should return that provider
	testCase(t, func(c *testContext) {
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-available",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: true,
		})
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-unavailable",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: false,
		})
		availableTools, notAvailableTools := Discover(config.New(), nil)
		t.Run("With one available provider returns that provider", func(t *testing.T) {
			assert.Len(t, availableTools, 1, "expected one available provider to be registered")
			assert.Equal(t, "provider-available", availableTools[0].Attributes().Name(),
				"expected the available provider to be returned")
			assert.Len(t, notAvailableTools, 1, "expected one not available provider to be registered")
		})
	})
}

func TestDiscoverMarshalling(t *testing.T) {
	testCase(t, func(c *testContext) {
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-one",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: true,
		})
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-two",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: true,
		})
		availableTools, notAvailableTools := Discover(config.New(), nil)
		bytes, err := json.Marshal(availableTools)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `[{"description":"Test Provider","name":"provider-one","reason":""},{"description":"Test Provider","name":"provider-two","reason":""}]`, string(bytes),
				"expected JSON to match the expected format")
		})
		assert.Empty(t, notAvailableTools, "expected no not available tools to be returned when no providers are registered")
	})
}

func TestGetDefaultPolicies(t *testing.T) {
	testCase(t, func(c *testContext) {
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-one",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: true,
			Policies: map[string]any{
				"provider-one-policy": "provider-one-policy-value",
			},
		})
		Register(&TestProvider{
			BasicToolsProvider: BasicToolsProvider{
				BasicToolsAttributes: BasicToolsAttributes{
					BasicFeatureAttributes: api.BasicFeatureAttributes{
						FeatureName:        "provider-two",
						FeatureDescription: "Test Provider",
					},
				},
			},
			Available: true,
			Policies: map[string]any{
				"provider-two-policy": "provider-two-policy-value",
			},
		})
		t.Run("GetDefaultPolicies returns expected policies", func(t *testing.T) {
			policies := GetDefaultPolicies()
			fmt.Printf("policies: %+v\n", policies)
			assert.Equal(t, map[string]any{"provider-one-policy": "provider-one-policy-value"}, policies["provider-one"], "expected the provider-one policy to be returned")
			assert.Equal(t, map[string]any{"provider-two-policy": "provider-two-policy-value"}, policies["provider-two"], "expected the provider-two policy to be returned")
		})
	})
}
