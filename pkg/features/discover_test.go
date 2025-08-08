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
	Reason    string
}

func (t *InferenceProvider) Attributes() inference.Attributes {
	return inference.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: t.Name,
		},
	}
}

func (t *InferenceProvider) Data() inference.Data {
	return inference.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: t.Reason,
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
		features := Discover(context.Background(), config.New())
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected an inference to be returned")
		})
		t.Run("With one available provider Inferences has one provider", func(t *testing.T) {
			assert.Len(t, features.Inferences, 1, "expected one inference provider to be returned")
			assert.Equal(t, "availableProvider", features.Inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
		t.Run("With one available provider Inference is set to that provider", func(t *testing.T) {
			assert.Equal(t, "availableProvider", (*features.Inference).Attributes().Name(),
				"expected the available provider to be returned")
		})
	})
}

func TestDiscoverKnownExplicitInference(t *testing.T) {
	// With one available inference provider, it should return that provider when specified in the config
	testCase(t, func(c *testContext) {
		inference.Register(&InferenceProvider{Name: "availableProvider", Available: true})
		inference.Register(&InferenceProvider{Name: "unavailableProvider", Available: false})
		cfg := config.New()
		cfg.Inference = func(s string) *string {
			return &s
		}("availableProvider")
		features := Discover(context.Background(), cfg)
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected an inference to be returned")
		})
		t.Run("With one available provider Inferences has one provider", func(t *testing.T) {
			assert.Len(t, features.Inferences, 1, "expected one inference provider to be returned")
			assert.Equal(t, "availableProvider", features.Inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
		t.Run("With one available provider Inference is set to that provider", func(t *testing.T) {
			assert.Equal(t, "availableProvider", (*features.Inference).Attributes().Name(),
				"expected the available provider to be returned")
		})
	})
}

func TestDiscoverUnknownExplicitInference(t *testing.T) {
	// With one available inference provider, it should not return another provider specified in the config
	testCase(t, func(c *testContext) {
		inference.Register(&InferenceProvider{Name: "availableProvider", Available: true})
		inference.Register(&InferenceProvider{Name: "unavailableProvider", Available: false})
		cfg := config.New()
		cfg.Inference = func(s string) *string {
			return &s
		}("otherProvider")
		features := Discover(context.Background(), cfg)
		t.Run("With one available provider returns features", func(t *testing.T) {
			assert.NotNil(t, features, "expected an inference to be returned")
		})
		t.Run("With one available provider Inferences has one provider", func(t *testing.T) {
			assert.Len(t, features.Inferences, 1, "expected no inference provider to be returned")
			assert.Equal(t, "availableProvider", features.Inferences[0].Attributes().Name(),
				"expected the available provider to be returned")
		})
		t.Run("With one available provider Inference is not set to unknown provider", func(t *testing.T) {
			assert.Nil(t, features.Inference,
				"expected no available provider to be returned")
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
		features := Discover(context.Background(), config.New())
		bytes, err := json.Marshal(features)
		t.Run("Marshalling returns no error", func(t *testing.T) {
			assert.Nil(t, err, "expected no error when marshalling inferences")
		})
		t.Run("Marshalling returns expected JSON", func(t *testing.T) {
			assert.JSONEq(t, `{"inference":{"local":false,"models":["gemini-2.0-flash"],"name":"gemini","public":true,"reason":"GEMINI_API_KEY is set"},"inferences":[{"local":false,"models":["gemini-2.0-flash"],"name":"gemini","public":true,"reason":"GEMINI_API_KEY is set"}],"inferencesNotAvailable":[{"local":true,"models":null,"name":"ollama","public":false,"reason":"http://localhost:11434 is not accessible"}],"tools":[{"name":"fs","reason":"filesystem is accessible"}],"toolsNotAvailable":null}`,
				string(bytes),
				"expected JSON to match the expected format")
		})
	})
}
