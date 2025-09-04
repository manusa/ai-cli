package config

import (
	"os"

	"github.com/manusa/ai-cli/pkg/api"
)

// ToolsConfig Configuration for tools
type ToolsConfig struct {
	// Provider ToolParameters specific for a provider
	Provider map[string]api.ToolsParameters `toml:"provider,omitempty"`
	// ToolParameters Global parameters for all tools
	api.ToolsParameters
}

type Config struct {
	// TODO: Should be moved to a separate InferenceConfig (same as is done for Tools -> ToolsConfig)
	Inference *string // An inference to use, if not set, the best inference will be used
	Model     *string // A model to use, if not set, the best model will be used

	toolsConfig ToolsConfig `toml:"tools,omitempty"`

	googleApiKey string // TODO: will likely be removed
	geminiModel  string // TODO: will likely be removed
}

func New() *Config {
	return &Config{
		googleApiKey: os.Getenv("GEMINI_API_KEY"),
		geminiModel:  "gemini-2.0-flash",
		toolsConfig: ToolsConfig{
			Provider: make(map[string]api.ToolsParameters),
		},
	}
}

func (c *Config) GoogleApiKey() string {
	return c.googleApiKey
}

func (c *Config) GeminiModel() string {
	return c.geminiModel
}

// ToolsParameters returns the merged tool configuration parameters for a specific tool
//
// TODO: moved this here from the tools package. IMO this should be maintained in this package which should hold
//
//	the single source of truth for configuration
//	IMO the workflow for configuration should be:
//	1) config.New creates a new Config with the default values (there's a test in the config package to ensure a spec)
//	2) The default config can be overridden by the user (Either by providing a partial config file -config.Read-
//	   or by using cmd flags -cmd.DiscoverCmdOptions-
//	3) The merged configuration is restricted/enforced by the policies (feature.Discover +
func (c *Config) ToolsParameters(toolName string) api.ToolsParameters {
	mergedParameters := api.ToolsParameters{
		Enabled: ptr(true),
		// TODO: all parameters are set to false by default, do we want to change this?
		// By default, tools are destructive and read-write
		ReadOnly:           ptr(false),
		DisableDestructive: ptr(false),
	}
	mergeableParameters := []api.ToolsParameters{c.toolsConfig.ToolsParameters}
	if toolParams, ok := c.toolsConfig.Provider[toolName]; ok {
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
	// Global policies override Global configurations
	c.toolsConfig.ToolsParameters = mergeToolsPolicies(policies.Tools.ToolsProviderPolicies, c.toolsConfig.ToolsParameters)

	// Global policies override provider-specific configuration
	for providerName, providerConfig := range c.toolsConfig.Provider {
		c.toolsConfig.Provider[providerName] = mergeToolsPolicies(policies.Tools.ToolsProviderPolicies, providerConfig)
	}

	// Provider-specific policies override or add provider-specific configuration
	for providerName, providerPolicies := range policies.Tools.Provider {
		originalParams, ok := c.toolsConfig.Provider[providerName]
		if !ok {
			originalParams = api.ToolsParameters{}
		}
		c.toolsConfig.Provider[providerName] = mergeToolsPolicies(providerPolicies, originalParams)
	}
}

func mergeToolsPolicies(toolsPolicies api.ToolsProviderPolicies, toolsParameters api.ToolsParameters) api.ToolsParameters {
	if toolsPolicies.Enabled != nil {
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

func ptr[T any](v T) *T {
	return &v
}
