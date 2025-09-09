package config

import (
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/stretchr/testify/suite"
)

type ConfigEnforceTestSuite struct {
	suite.Suite
	baseConfig *Config
}

func (s *ConfigEnforceTestSuite) SetupTest() {
	s.baseConfig = New()
	s.baseConfig.toolsConfig.Enabled = ptr(true)
	s.baseConfig.toolsConfig.ReadOnly = ptr(false)
	s.baseConfig.toolsConfig.DisableDestructive = ptr(false)
	s.baseConfig.toolsConfig.Provider["existing-provider-all-false"] = api.ToolsParameters{
		Enabled:            ptr(false),
		ReadOnly:           ptr(false),
		DisableDestructive: ptr(false),
	}
	s.baseConfig.toolsConfig.Provider["existing-provider-all-true"] = api.ToolsParameters{
		Enabled:            ptr(true),
		ReadOnly:           ptr(true),
		DisableDestructive: ptr(true),
	}
}

func (s *ConfigEnforceTestSuite) TestNilOrEmptyPolicies() {
	testCases := []*api.Policies{nil, test.Must(policies.ReadToml(""))}
	for _, tc := range testCases {
		s.baseConfig.Enforce(tc)
		s.Run("global config remains unchanged", func() {
			s.Equal(ptr(true), s.baseConfig.toolsConfig.Enabled, "Expected Enabled to remain true")
			s.Equal(ptr(false), s.baseConfig.toolsConfig.ReadOnly, "Expected ReadOnly to remain false")
			s.Equal(ptr(false), s.baseConfig.toolsConfig.DisableDestructive, "Expected DisableDestructive to remain false")
		})
		s.Run("provider-specific config remains unchanged", func() {
			params := s.baseConfig.ToolsParameters("existing-provider-all-false")
			s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to remain false")
			s.Equal(ptr(false), params.ReadOnly, "Expected provider ReadOnly to remain false")
			s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to remain false")

			params = s.baseConfig.ToolsParameters("existing-provider-all-true")
			s.Equal(ptr(true), params.Enabled, "Expected provider Enabled to remain true")
			s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to remain true")
			s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to remain true")
		})
	}
}

func (s *ConfigEnforceTestSuite) TestGlobalPolicies() {
	p := test.Must(policies.ReadToml(`
[tools]
enabled = false
read-only = true
non-destructive = true
`))
	s.baseConfig.Enforce(p)
	s.Run("global policies override config", func() {
		s.Equal(ptr(false), s.baseConfig.toolsConfig.Enabled, "Expected Enabled to be false as per policies")
		s.Equal(ptr(true), s.baseConfig.toolsConfig.ReadOnly, "Expected ReadOnly to be true as per policies")
		s.Equal(ptr(true), s.baseConfig.toolsConfig.DisableDestructive, "Expected DisableDestructive to be true as per policies")
	})
	s.Run("global policies override provider-specific config", func() {
		params := s.baseConfig.ToolsParameters("existing-provider-all-false")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per global policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per global policies")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per global policies")

		params = s.baseConfig.ToolsParameters("existing-provider-all-true")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per global policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per global policies")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per global policies")
	})
}

func (s *ConfigEnforceTestSuite) TestProviderPolicies() {
	p := test.Must(policies.ReadToml(`
[tools.provider.existing-provider-all-false]
enabled = true
read-only = true
non-destructive = true

[tools.provider.existing-provider-all-true]
enabled = false
read-only = false
non-destructive = false
`))
	s.baseConfig.Enforce(p)
	s.Run("global config remains unchanged", func() {
		s.Equal(ptr(true), s.baseConfig.toolsConfig.Enabled, "Expected Enabled to remain true")
		s.Equal(ptr(false), s.baseConfig.toolsConfig.ReadOnly, "Expected ReadOnly to remain false")
		s.Equal(ptr(false), s.baseConfig.toolsConfig.DisableDestructive, "Expected DisableDestructive to remain false")
	})
	s.Run("provider-specific policies override config", func() {
		params := s.baseConfig.ToolsParameters("existing-provider-all-false")
		s.Equal(ptr(true), params.Enabled, "Expected provider Enabled to be true as per policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per policies")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per policies")

		params = s.baseConfig.ToolsParameters("existing-provider-all-true")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per policies")
		s.Equal(ptr(false), params.ReadOnly, "Expected provider ReadOnly to be false as per policies")
		s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to be false as per policies")
	})
}

func (s *ConfigEnforceTestSuite) TestMixedPolicies() {
	p := test.Must(policies.ReadToml(`
[tools]
enabled = false

[tools.provider.existing-provider-all-false]
read-only = true

[tools.provider.existing-provider-all-true]
enabled = false
non-destructive = true
`))
	s.baseConfig.Enforce(p)
	s.Run("global policies override config", func() {
		s.Equal(ptr(false), s.baseConfig.toolsConfig.Enabled, "Expected Enabled to be false as per policies")
		s.Equal(ptr(false), s.baseConfig.toolsConfig.ReadOnly, "Expected ReadOnly to remain false")
		s.Equal(ptr(false), s.baseConfig.toolsConfig.DisableDestructive, "Expected DisableDestructive to remain false")
	})
	s.Run("provider-specific policies override config", func() {
		params := s.baseConfig.ToolsParameters("existing-provider-all-false")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per global policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per provider policies")
		s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to remain false")

		params = s.baseConfig.ToolsParameters("existing-provider-all-true")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per provider policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to remain true")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per provider policies")
	})
}

func (s *ConfigEnforceTestSuite) TestMixedPoliciesWithNewProvider() {
	p := test.Must(policies.ReadToml(`
[tools]
enabled = false
read-only = true
non-destructive = true

[tools.provider.new-provider]
enabled = true
read-only = false
non-destructive = false
`))
	s.baseConfig.Enforce(p)
	s.Run("global policies override config", func() {
		s.Equal(ptr(false), s.baseConfig.toolsConfig.Enabled, "Expected Enabled to be false as per policies")
		s.Equal(ptr(true), s.baseConfig.toolsConfig.ReadOnly, "Expected ReadOnly to be true as per policies")
		s.Equal(ptr(true), s.baseConfig.toolsConfig.DisableDestructive, "Expected DisableDestructive to be true as per policies")
	})
	s.Run("new provider-specific policies are applied", func() {
		params := s.baseConfig.ToolsParameters("new-provider")
		s.Equal(ptr(true), params.Enabled, "Expected provider Enabled to be true as per policies")
		s.Equal(ptr(false), params.ReadOnly, "Expected provider ReadOnly to be false as per policies")
		s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to be false as per policies")
	})
}

func (s *ConfigEnforceTestSuite) TestMixedPoliciesToolsParameters() {
	p := test.Must(policies.ReadToml(`
[tools]
enabled = false
read-only = true
non-destructive = true

[tools.provider.existing-provider-all-false]
read-only = false

[tools.provider.existing-provider-all-true]
enabled = false
non-destructive = false

[tools.provider.new-provider]
enabled = true
read-only = false
non-destructive = false
`))
	s.baseConfig.Enforce(p)
	s.Run("ToolsParameters with non-existing provider returns overridden config", func() {
		params := s.baseConfig.ToolsParameters("non-existing-provider")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per global policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per global policies")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per global policies")
	})
	s.Run("ToolsParameters with existing provider and partial policies returns merged config", func() {
		params := s.baseConfig.ToolsParameters("existing-provider-all-false")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per global policies")
		s.Equal(ptr(false), params.ReadOnly, "Expected provider ReadOnly to be false as per provider policies")
		s.Equal(ptr(true), params.DisableDestructive, "Expected provider DisableDestructive to be true as per global policies")

		params = s.baseConfig.ToolsParameters("existing-provider-all-true")
		s.Equal(ptr(false), params.Enabled, "Expected provider Enabled to be false as per provider policies")
		s.Equal(ptr(true), params.ReadOnly, "Expected provider ReadOnly to be true as per global policies")
		s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to be false as per provider policies")
	})
	s.Run("ToolsParameters with new provider returns overridden config", func() {
		params := s.baseConfig.ToolsParameters("new-provider")
		s.Equal(ptr(true), params.Enabled, "Expected provider Enabled to be true as per policies")
		s.Equal(ptr(false), params.ReadOnly, "Expected provider ReadOnly to be false as per policies")
		s.Equal(ptr(false), params.DisableDestructive, "Expected provider DisableDestructive to be false as per policies")
	})
}

func TestConfigEnforce(t *testing.T) {
	suite.Run(t, new(ConfigEnforceTestSuite))
}
