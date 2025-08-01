package features

import (
	"context"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/config"
)

type testContext struct {
}

func (c *testContext) beforeEach(t *testing.T) {
	t.Helper()
	inference.Clear()
	tools.Clear()
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

type InferenceProvider struct {
	Name      string
	Available bool
}

func (t *InferenceProvider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: t.Name,
		},
	}
}

func (t *InferenceProvider) IsAvailable(_ *config.Config) bool {
	return t.Available
}

func (t *InferenceProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

func TestDiscoverInference(t *testing.T) {
	// With no inference providers registered, it should return an error
	testCase(t, func(c *testContext) {
		t.Run("With no inference providers registered returns an error", func(t *testing.T) {
			_, err := Discover(config.New())
			assert.NotNil(t, err, "expected an error when no providers are registered")
		})
	})
	// With one available inference provider, it should return that provider
	testCase(t, func(c *testContext) {
		inference.Register(&InferenceProvider{Name: "availableProvider", Available: true})
		inference.Register(&InferenceProvider{Name: "unavailableProvider", Available: false})
		features, err := Discover(config.New())
		t.Run("With one available provider has no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error")
		})
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected an inference to be returned")
		})
		t.Run("With one available provider Inferences has one provider", func(t *testing.T) {
			assert.Len(t, features.Inferences, 1, "expected one inference provider to be returned")
			assert.Equal(t, "availableProvider", features.Inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
		t.Run("With one available provider Inference is set to that provider", func(t *testing.T) {
			assert.Equal(t, "availableProvider", features.Inference.Attributes().Name(),
				"expected the available provider to be returned")
		})
	})
}
