package config

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/test"
	"github.com/stretchr/testify/suite"
)

type IsToolsProviderEnabledTestSuite struct {
	suite.Suite
	baseConfig *Config
}

func (s *IsToolsProviderEnabledTestSuite) SetupTest() {
	s.baseConfig = New()
}

func (s *IsToolsProviderEnabledTestSuite) TestEmptyPolicies() {
	s.baseConfig.Enforce(test.Must(policies.ReadToml("")))
	s.Run("enabled by default", func() {
		result := s.baseConfig.IsToolsProviderEnabled(test.NewToolsProvider("provider"))
		s.Equal(true, result, "Expected tools provider to be enabled by default")
	})
}

func (s *IsToolsProviderEnabledTestSuite) TestPoliciesGlobalDisable() {
	s.baseConfig.toolsConfig.Enabled = ptr(true)
	s.baseConfig.toolsConfig.Provider["provider-enabled-in-config"] = api.ToolsParameters{Enabled: ptr(true)}
	s.baseConfig.Enforce(test.Must(policies.ReadToml(`
[tools]
enabled = false

[tools.provider.provider-enabled]
enabled = true
`)))
	s.Run("disabled by default", func() {
		result := s.baseConfig.IsToolsProviderEnabled(test.NewToolsProvider("provider-disabled"))
		s.Equal(false, result, "Expected tools provider to be disabled by default")
	})
	s.Run("enabled by name", func() {
		result := s.baseConfig.IsToolsProviderEnabled(test.NewToolsProvider("provider-enabled"))
		s.Equal(true, result, "Expected tools provider to be enabled by name")
	})
	s.Run("disables providers enabled in config", func() {
		result := s.baseConfig.IsToolsProviderEnabled(test.NewToolsProvider("provider-enabled-in-config"))
		s.Equal(false, result, "Expected tools provider to be disabled by config")
	})
}

func TestIsToolsProviderEnabledTestSuite(t *testing.T) {
	suite.Run(t, new(IsToolsProviderEnabledTestSuite))
}
