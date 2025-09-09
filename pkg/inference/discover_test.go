package inference

import (
	"context"
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/stretchr/testify/suite"
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
		Register(test.NewInferenceProvider(
			"testProvider",
			test.WithInferenceAvailable(),
			test.WithInferenceLocal(),
		))
		s.Contains(providers, "testProvider",
			"expected provider %s to be registered in the providers %v", "testProvider", providers)
	})
	// Registering a provider with the same name should panic
	s.Run("Registering a provider with the same name panics", func() {
		provider := test.NewInferenceProvider(
			"duplicateProvider",
			test.WithInferenceAvailable(),
			test.WithInferenceLocal(),
		)
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
	provider := test.NewInferenceProvider("the-provider")
	Register(provider)
	Initialize(context.Background())
	s.Run("Initialize calls Initialize on all providers", func() {
		s.True(provider.Initialized, "expected provider to be initialized")
	})
}

func TestDiscover(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
