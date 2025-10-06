package config

import (
	"github.com/manusa/ai-cli/pkg/api"
)

const (
	DefaultInferenceEnabled = true
)

type InferenceConfig struct {
	Inference *string // An inference to use, if not set, the best inference will be used
	// Provider InferenceParameters specific for a provider
	Provider map[string]api.InferenceParameters `toml:"provider,omitempty"`
	// InferenceParameters Global parameters for all tools
	api.InferenceParameters
}

// ToolsConfig Configuration for tools
type ToolsConfig struct {
	// Provider ToolParameters specific for a provider
	Provider map[string]api.ToolsParameters `toml:"provider,omitempty"`
	// ToolParameters Global parameters for all tools
	api.ToolsParameters
}

type Config struct {
	InferenceConfig InferenceConfig `toml:"inferences,omitempty"`
	toolsConfig     ToolsConfig     `toml:"tools,omitempty"`

	policies *api.Policies // TODO: should be removed in favor of ToolsConfig and InferenceConfig above
}

// New creates a new configuration with defaults
//
//	TBD: The workflow for configuration should be:
//	1) config.New creates a new Config with the default values (there's a test in the config package to ensure a spec)
//	2) The default config can be overridden by the user (Either by providing a partial config file -config.Read-
//	   or by using cmd flags -cmd.DiscoverCmdOptions-)
//	3) The merged configuration is restricted/enforced by the policies Config.Enforce
func New() *Config {
	return &Config{
		InferenceConfig: InferenceConfig{
			InferenceParameters: api.InferenceParameters{
				// By default, all inference providers are enabled
				Enabled: ptr(true),
			},
			Provider: make(map[string]api.InferenceParameters),
		},
		toolsConfig: ToolsConfig{
			ToolsParameters: api.ToolsParameters{
				Enabled: ptr(true),
				// TODO: all parameters are set to false by default, do we want to change this?
				// By default, tools are destructive and read-write
				ReadOnly:           ptr(false),
				DisableDestructive: ptr(false),
			},
			Provider: make(map[string]api.ToolsParameters),
		},
	}
}

func (c *Config) Inference() *string {
	return c.InferenceConfig.Inference
}

// InferenceParameters returns the merged inference configuration parameters for a specific provider
// It considers both the global configuration and the provider-specific configuration
// Provider-specific configuration takes precedence over global configuration
func (c *Config) InferenceParameters(inferenceProviderName string) api.InferenceParameters {
	mergedParameters := api.InferenceParameters{}
	mergeableParameters := []api.InferenceParameters{c.InferenceConfig.InferenceParameters}
	if providerParams, ok := c.InferenceConfig.Provider[inferenceProviderName]; ok {
		mergeableParameters = append(mergeableParameters, providerParams)
	}
	// Merge configurations by precedence
	for _, params := range mergeableParameters {
		if params.Enabled != nil {
			mergedParameters.Enabled = params.Enabled
		}
	}
	return mergedParameters
}

// ToolsParameters returns the merged tool configuration parameters for a specific tool
// It considers both the global configuration and the provider-specific configuration
// Provider-specific configuration takes precedence over global configuration
func (c *Config) ToolsParameters(toolProviderName string) api.ToolsParameters {
	mergedParameters := api.ToolsParameters{}
	mergeableParameters := []api.ToolsParameters{c.toolsConfig.ToolsParameters}
	if toolParams, ok := c.toolsConfig.Provider[toolProviderName]; ok {
		mergeableParameters = append(mergeableParameters, toolParams)
	}
	// Merge configurations by precedence
	for _, params := range mergeableParameters {
		if params.Enabled != nil {
			mergedParameters.Enabled = params.Enabled
		}
		if params.ReadOnly != nil {
			mergedParameters.ReadOnly = params.ReadOnly
		}
		if params.DisableDestructive != nil {
			mergedParameters.DisableDestructive = params.DisableDestructive
		}
	}
	return mergedParameters
}

func (c *Config) Enforce(policies *api.Policies) {
	if policies == nil {
		return
	}
	c.policies = policies // TODO: should be removed in favor of ToolsConfig and *InferenceConfig*
	// Global policies override Global configurations
	c.InferenceConfig.InferenceParameters = mergeInferencesPolicies(policies.Inferences.InferenceProviderPolicies, c.InferenceConfig.InferenceParameters)
	c.toolsConfig.ToolsParameters = mergeToolsPolicies(policies.Tools.ToolsProviderPolicies, c.toolsConfig.ToolsParameters)

	// Global policies override provider-specific configuration
	for providerName, providerParameters := range c.InferenceConfig.Provider {
		c.InferenceConfig.Provider[providerName] = mergeInferencesPolicies(policies.Inferences.InferenceProviderPolicies, providerParameters)
	}
	for providerName, providerConfig := range c.toolsConfig.Provider {
		c.toolsConfig.Provider[providerName] = mergeToolsPolicies(policies.Tools.ToolsProviderPolicies, providerConfig)
	}

	// Provider-specific policies override or add provider-specific configuration
	for providerName, providerPolicies := range policies.Inferences.Provider {
		originalParams, ok := c.InferenceConfig.Provider[providerName]
		if !ok {
			originalParams = api.InferenceParameters{}
		}
		c.InferenceConfig.Provider[providerName] = mergeInferencesPolicies(providerPolicies, originalParams)
	}
	for providerName, providerPolicies := range policies.Tools.Provider {
		originalParams, ok := c.toolsConfig.Provider[providerName]
		if !ok {
			originalParams = api.ToolsParameters{}
		}
		c.toolsConfig.Provider[providerName] = mergeToolsPolicies(providerPolicies, originalParams)
	}

}

func mergeInferencesPolicies(inferencesPolicies api.InferenceProviderPolicies, inferenceParameters api.InferenceParameters) api.InferenceParameters {
	if inferencesPolicies.Enabled != nil {
		// TODO there might be issues here in case policy enables a tool that's disabled by config. We need to evaluate this case specifically.
		inferenceParameters.Enabled = inferencesPolicies.Enabled
	}
	return inferenceParameters
}

func mergeToolsPolicies(toolsPolicies api.ToolsProviderPolicies, toolsParameters api.ToolsParameters) api.ToolsParameters {
	if toolsPolicies.Enabled != nil {
		// TODO there might be issues here in case policy enables a tool that's disabled by config. We need to evaluate this case specifically.
		toolsParameters.Enabled = toolsPolicies.Enabled
	}
	if toolsPolicies.ReadOnly != nil {
		toolsParameters.ReadOnly = toolsPolicies.ReadOnly
	}
	if toolsPolicies.NonDestructive != nil {
		toolsParameters.DisableDestructive = toolsPolicies.NonDestructive
	}
	//nolint:staticcheck
	if toolsPolicies.Local != nil {
		// TODO: I don't understand what policies.Tools.Local is meant for
	}
	return toolsParameters
}

func (c *Config) IsInferenceProviderEnabled(feature api.Feature[api.InferenceAttributes]) bool {
	// TODO: relying only on *c.InferenceParameters(feature.Attributes().Name()).Enabled
	//       won't work even if we consider policies.
	//       It becomes especially hard for the scenario:
	//       "globally disabled and enabled by property remote, preserves remote providers enabled"
	//       Since the provider has been disabled in the Enforce step, we cannot know if it was
	//       disabled by global policy, provider-specific policy, or configuration.
	//       This means that is very hard to know if the provider should be re-enabled or preserved disabled.

	// TODO: considering the previous comment, for now, this method is not checking for config-disabled inferences
	if c.policies == nil {
		return DefaultInferenceEnabled
	}
	providerName := feature.Attributes().Name()
	if c.policies.Inferences.Provider[providerName].Enabled != nil {
		return *c.policies.Inferences.Provider[providerName].Enabled
	}

	providerLocal := feature.Attributes().Local()
	if c.policies.Inferences.Property.Remote.Enabled != nil {
		if !*c.policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return false
		}
		if *c.policies.Inferences.Property.Remote.Enabled && !providerLocal {
			return true
		}
	}

	if c.policies.Inferences.Enabled != nil {
		return *c.policies.Inferences.Enabled
	}

	return DefaultInferenceEnabled
}

func (c *Config) IsToolsProviderEnabled(feature api.Feature[api.ToolsAttributes]) bool {
	// TODO should be disabled if read-only/... is set by policy and provider does not support it
	return *c.ToolsParameters(feature.Attributes().Name()).Enabled
}

func ptr[T any](v T) *T {
	return &v
}
