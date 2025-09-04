package config

import (
	"testing"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/stretchr/testify/suite"
)

type ConfigToolsParametersTestSuite struct {
	suite.Suite
}

func (s *ConfigToolsParametersTestSuite) TestWithDefaultParameters() {
	defaultCfg := New()
	s.Run("With empty tool name", func() {
		result := defaultCfg.ToolsParameters("")
		s.Equal(ptr(true), result.Enabled, "Expected Enabled to be true by default")
		s.Equal(ptr(false), result.ReadOnly, "Expected ReadOnly to be false by default")
		s.Equal(ptr(false), result.DisableDestructive, "Expected DisableDestructive to be false by default")
	})
	s.Run("With non-existing tool name", func() {
		result := defaultCfg.ToolsParameters("non-existing-tool")
		s.Equal(ptr(true), result.Enabled, "Expected Enabled to be true by default")
		s.Equal(ptr(false), result.ReadOnly, "Expected ReadOnly to be false by default")
		s.Equal(ptr(false), result.DisableDestructive, "Expected DisableDestructive to be false by default")
	})
	cfgWithProvider := New()
	cfgWithProvider.toolsConfig.Provider["existing-provider"] = api.ToolsParameters{}
	s.Run("With tool providers and existing tool name and empty parameters", func() {
		result := defaultCfg.ToolsParameters("existing-provider")
		s.Equal(ptr(true), result.Enabled, "Expected Enabled to be true by default")
		s.Equal(ptr(false), result.ReadOnly, "Expected ReadOnly to be false by default")
		s.Equal(ptr(false), result.DisableDestructive, "Expected DisableDestructive to be false by default")
	})
}

func (s *ConfigToolsParametersTestSuite) TestWithToolsConfig() {
	cfgWithToolsConfig := New()
	cfgWithToolsConfig.toolsConfig.Enabled = ptr(false)
	cfgWithToolsConfig.toolsConfig.ReadOnly = ptr(true)
	cfgWithToolsConfig.toolsConfig.DisableDestructive = ptr(true)
	s.Run("With empty tool name", func() {
		result := cfgWithToolsConfig.ToolsParameters("")
		s.Equal(ptr(false), result.Enabled, "Expected Enabled to be false as per global config")
		s.Equal(ptr(true), result.ReadOnly, "Expected ReadOnly to be true as per global config")
		s.Equal(ptr(true), result.DisableDestructive, "Expected DisableDestructive to be true as per global config")
	})
	s.Run("With non-existing tool name", func() {
		result := cfgWithToolsConfig.ToolsParameters("non-existing-tool")
		s.Equal(ptr(false), result.Enabled, "Expected Enabled to be false as per global config")
		s.Equal(ptr(true), result.ReadOnly, "Expected ReadOnly to be true as per global config")
		s.Equal(ptr(true), result.DisableDestructive, "Expected DisableDestructive to be true as per global config")
	})
	cfgWithProviderEmpty := *cfgWithToolsConfig
	cfgWithProviderEmpty.toolsConfig.Provider["existing-provider"] = api.ToolsParameters{}
	s.Run("With tool providers and existing tool name and empty parameters", func() {
		result := cfgWithProviderEmpty.ToolsParameters("existing-provider")
		s.Equal(ptr(false), result.Enabled, "Expected Enabled to be false as per global config")
		s.Equal(ptr(true), result.ReadOnly, "Expected ReadOnly to be true as per global config")
		s.Equal(ptr(true), result.DisableDestructive, "Expected DisableDestructive to be true as per global config")
	})
	cfgWithProvider := *cfgWithToolsConfig
	cfgWithProvider.toolsConfig.Provider["existing-provider"] = api.ToolsParameters{
		Enabled:            ptr(true),
		ReadOnly:           ptr(false),
		DisableDestructive: ptr(false),
	}
	s.Run("With tool providers and existing tool name and full parameters", func() {
		result := cfgWithProvider.ToolsParameters("existing-provider")
		s.Equal(ptr(true), result.Enabled, "Expected Enabled to be true as per provider config")
		s.Equal(ptr(false), result.ReadOnly, "Expected ReadOnly to be false as per provider config")
		s.Equal(ptr(false), result.DisableDestructive, "Expected DisableDestructive to be false as per provider config")
	})
	s.Run("With tool providers and non-existing tool name", func() {
		result := cfgWithProvider.ToolsParameters("non-existing-tool")
		s.Equal(ptr(false), result.Enabled, "Expected Enabled to be false as per global config")
		s.Equal(ptr(true), result.ReadOnly, "Expected ReadOnly to be true as per global config")
		s.Equal(ptr(true), result.DisableDestructive, "Expected DisableDestructive to be true as per global config")
	})
}

func TestConfigToolsParameters(t *testing.T) {
	suite.Run(t, new(ConfigToolsParametersTestSuite))
}
