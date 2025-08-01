package features

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/inference/gemini"
	"github.com/manusa/ai-cli/pkg/inference/ollama"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/fs"
	"github.com/stretchr/testify/assert"
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

func (t *InferenceProvider) GetModels(_ context.Context, _ *config.Config) ([]string, error) {
	return []string{}, nil
}

func (t *InferenceProvider) IsAvailable(_ *config.Config) bool {
	return t.Available
}

func (t *InferenceProvider) GetInference(_ context.Context, _ *config.Config) (model.ToolCallingChatModel, error) {
	return nil, nil
}

func (t *InferenceProvider) MarshalJSON() ([]byte, error) { return nil, nil }

func TestDiscoverInference(t *testing.T) {
	// With one available inference provider, it should return that provider
	testCase(t, func(c *testContext) {
		inference.Register(&InferenceProvider{Name: "availableProvider", Available: true})
		inference.Register(&InferenceProvider{Name: "unavailableProvider", Available: false})
		features := Discover(config.New())
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

func TestDiscoverMarshal(t *testing.T) {
	testCase(t, func(c *testContext) {
		_ = os.Setenv("GEMINI_API_KEY", "test-key")
		t.Cleanup(func() { _ = os.Unsetenv("GEMINI_API_KEY") })
		inference.Register(&gemini.Provider{})
		inference.Register(&ollama.Provider{})
		tools.Register(&fs.Provider{})
		features := Discover(config.New())
		bytes, err := json.Marshal(features)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `{"inference":{"name":"gemini","Distant":true},"inferences":[{"name":"gemini","Distant":true}],"tools":[{"name":"fs"}]}`,
				string(bytes),
				"expected JSON to match the expected format")
		})
	})
}
