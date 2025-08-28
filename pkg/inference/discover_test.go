package inference

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

type TestProvider struct {
	BasicInferenceProvider
	Available bool
}

func (t *TestProvider) IsAvailable(_ *config.Config, _ any) bool {
	return t.Available
}

func (t *TestProvider) GetDefaultPolicies() map[string]any {
	return nil
}

func (t *TestProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
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
				BasicInferenceProvider: BasicInferenceProvider{
					BasicInferenceAttributes: BasicInferenceAttributes{
						api.BasicFeatureAttributes{FeatureName: "testProvider", FeatureDescription: "Test Provider"},
						true,
						false,
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
				BasicInferenceProvider: BasicInferenceProvider{
					BasicInferenceAttributes: BasicInferenceAttributes{
						api.BasicFeatureAttributes{FeatureName: "duplicateProvider", FeatureDescription: "Test Provider"},
						true,
						false,
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
			availableInferences, notAvailableInferences := Discover(config.New(), nil)
			assert.Empty(t, availableInferences, "expected no inferences to be returned when no providers are registered")
			assert.Empty(t, notAvailableInferences, "expected no not available inferences to be returned when no providers are registered")
		})
	})
	// With one available provider, it should return that provider
	testCase(t, func(c *testContext) {
		Register(&TestProvider{
			BasicInferenceProvider: BasicInferenceProvider{
				BasicInferenceAttributes: BasicInferenceAttributes{
					api.BasicFeatureAttributes{FeatureName: "provider-available", FeatureDescription: "Test Provider"},
					true,
					false,
				},
			},
			Available: true,
		})
		Register(&TestProvider{
			BasicInferenceProvider: BasicInferenceProvider{
				BasicInferenceAttributes: BasicInferenceAttributes{
					api.BasicFeatureAttributes{FeatureName: "provider-unavailable", FeatureDescription: "Test Provider"},
					true,
					false,
				},
			},
			Available: false,
		})
		availableInferences, notAvailableInferences := Discover(config.New(), nil)
		t.Run("With one available provider returns that provider", func(t *testing.T) {
			assert.Len(t, availableInferences, 1, "expected one available provider to be registered")
			assert.Equal(t, "provider-available", availableInferences[0].Attributes().Name(),
				"expected the available provider to be returned")
			assert.Len(t, notAvailableInferences, 1, "expected one not available provider")
		})
	})
}
