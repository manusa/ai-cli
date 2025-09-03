package tools

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/test"
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

func TestDiscover(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
