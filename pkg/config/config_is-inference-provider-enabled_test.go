package config

import (
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/stretchr/testify/suite"
)

type IsInferenceProviderEnabledTestSuite struct {
	suite.Suite
	baseConfig *Config
}

func (s *IsInferenceProviderEnabledTestSuite) SetupTest() {
	s.baseConfig = New()
}

func (s *IsInferenceProviderEnabledTestSuite) TestEmptyPolicies() {
	s.baseConfig.Enforce(test.Must(policies.ReadToml("")))
	s.Run("enabled by default", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider"))
		s.Equal(true, result, "Expected inference provider to be enabled by default")
	})
}

func (s *IsInferenceProviderEnabledTestSuite) TestPoliciesGlobalDisable() {
	s.baseConfig.Enforce(test.Must(policies.ReadToml(`
[inferences]
enabled = false

[inferences.provider.provider-enabled]
enabled = true
`)))
	s.Run("disabled by default", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-disabled"))
		s.Equal(false, result, "Expected inference provider to be disabled by default")
	})
	s.Run("enabled by name", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-enabled"))
		s.Equal(true, result, "Expected inference provider to be enabled by name")
	})
}

func (s *IsInferenceProviderEnabledTestSuite) TestPoliciesRemote() {
	s.baseConfig.Enforce(test.Must(policies.ReadToml(`
[inferences.property.remote]
enabled = false
`)))
	s.Run("disabled by property remote, disables remote providers", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-remote"))
		s.Equal(false, result, "Expected inference provider to be disabled by property remote")
	})
	s.Run("disabled by property remote, preserves local providers enabled", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-local", test.WithInferenceLocal()))
		s.Equal(true, result, "Expected inference provider to be enabled by default")
	})
	s.baseConfig.Enforce(test.Must(policies.ReadToml(`
[inferences]
enabled = false

[inferences.property.remote]
enabled = true
`)))
	s.Run("globally disabled and enabled by property remote, preserves remote providers enabled", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-remote"))
		s.Equal(true, result, "Expected inference provider to be enabled by property remote")
	})
	s.Run("globally disabled and enabled by property remote, local provider is disabled", func() {
		result := s.baseConfig.IsInferenceProviderEnabled(test.NewInferenceProvider("provider-local", test.WithInferenceLocal()))
		s.Equal(false, result, "Expected inference provider to be enabled by default")
	})
}

func TestIsInferenceProviderEnabled(t *testing.T) {
	suite.Run(t, new(IsInferenceProviderEnabledTestSuite))
}
