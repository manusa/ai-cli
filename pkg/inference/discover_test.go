package inference

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/stretchr/testify/assert"
)

type TestProvider struct {
	Name      string
	Distant   bool
	Available bool
}

func (t *TestProvider) Attributes() Attributes {
	return Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: t.Name,
		},
		Distant: t.Distant,
	}
}

func (t *TestProvider) IsAvailable(_ *config.Config) bool {
	return t.Available
}

func (t *TestProvider) GetModels(_ context.Context, _ *config.Config) ([]string, error) {
	return []string{}, nil
}

func (t *TestProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

func (t *TestProvider) MarshalJSON() ([]byte, error) { return json.Marshal(t.Attributes()) }

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
			Register(&TestProvider{Name: "testProvider", Distant: false, Available: true})
			assert.Contains(t, providers, "testProvider",
				"expected provider %s to be registered in the providers %v", "testProvider", providers)
		})
	})
	// Registering a provider with the same name should panic
	testCase(t, func(c *testContext) {
		t.Run("Registering a provider with the same name panics", func(t *testing.T) {
			provider := &TestProvider{Name: "duplicateProvider", Distant: false, Available: true}
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
			inferences := Discover(config.New())
			assert.Empty(t, inferences, "expected no inferences to be returned when no providers are registered")
		})
	})
	// With one available provider, it should return that provider
	testCase(t, func(c *testContext) {
		Register(&TestProvider{Name: "availableProvider", Distant: false, Available: true})
		Register(&TestProvider{Name: "unavailableProvider", Distant: false, Available: false})
		inferences := Discover(config.New())
		t.Run("With one available provider returns that provider", func(t *testing.T) {
			assert.Len(t, inferences, 1, "expected one available provider to be registered")
			assert.Equal(t, "availableProvider", inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
	})
}

func TestDiscoverMarshalling(t *testing.T) {
	testCase(t, func(c *testContext) {
		Register(&TestProvider{Name: "provider-one", Distant: false, Available: true})
		Register(&TestProvider{Name: "provider-two", Distant: false, Available: true})
		inferences := Discover(config.New())
		bytes, err := json.Marshal(inferences)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `[{"name":"provider-one","distant":false},{"name":"provider-two","distant":false}]`, string(bytes),
				"expected JSON to match the expected format")
		})
	})
}
