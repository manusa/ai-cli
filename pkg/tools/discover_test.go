package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/stretchr/testify/assert"

	"github.com/manusa/ai-cli/pkg/config"
)

type TestProvider struct {
	Name      string
	Available bool
	Reason    string
	Tools     []*api.Tool
}

func (t *TestProvider) Attributes() Attributes {
	return Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: t.Name,
		},
	}
}

func (t *TestProvider) Data() Data {
	return Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: t.Reason,
		},
	}
}

func (t *TestProvider) IsAvailable(_ *config.Config) bool {
	return t.Available
}

func (t *TestProvider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return t.Tools, nil
}

func (t *TestProvider) MarshalJSON() ([]byte, error) { return json.Marshal(t.Attributes()) }

func (t *TestProvider) SetReason(reason string) {
	t.Reason = reason
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
			Register(&TestProvider{Name: "testProvider", Available: true})
			assert.Contains(t, providers, "testProvider",
				"expected provider %s to be registered in the providers %v", "testProvider", providers)
		})
	})
	// Registering a provider with the same name should panic
	testCase(t, func(c *testContext) {
		t.Run("Registering a provider with the same name panics", func(t *testing.T) {
			provider := &TestProvider{Name: "duplicateProvider", Available: true}
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
			availableTools, notAvailableTools := Discover(config.New())
			assert.Empty(t, availableTools, "expected no available tools to be returned when no providers are registered")
			assert.Empty(t, notAvailableTools, "expected no not available tools to be returned when no providers are registered")
		})
	})
	// With one available provider, it should return that provider
	testCase(t, func(c *testContext) {
		Register(&TestProvider{Name: "availableProvider", Available: true})
		Register(&TestProvider{Name: "unavailableProvider", Available: false})
		availableTools, notAvailableTools := Discover(config.New())
		t.Run("With one available provider returns that provider", func(t *testing.T) {
			assert.Len(t, availableTools, 1, "expected one available provider to be registered")
			assert.Equal(t, "availableProvider", availableTools[0].Attributes().Name(),
				"expected the available provider to be returned")
			assert.Len(t, notAvailableTools, 1, "expected one not available provider to be registered")
		})
	})
}

func TestDiscoverMarshalling(t *testing.T) {
	testCase(t, func(c *testContext) {
		Register(&TestProvider{Name: "provider-one", Available: true})
		Register(&TestProvider{Name: "provider-two", Available: true})
		availableTools, notAvailableTools := Discover(config.New())
		bytes, err := json.Marshal(availableTools)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `[{"name":"provider-one"},{"name":"provider-two"}]`, string(bytes),
				"expected JSON to match the expected format")
		})
		assert.Empty(t, notAvailableTools, "expected no not available tools to be returned when no providers are registered")
	})
}
