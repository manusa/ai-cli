package api

type InferencePropertyPolicies struct {
	Remote InferenceProviderPolicies `toml:"remote,omitempty"`
	Public InferenceProviderPolicies `toml:"public,omitempty"`
}

// Inference policies can be defined at different levels (highest priority first): by a provider name, by a provider property, or globally
// Policies are pointers, this means that if a policy is not set at a specific level, the value at a lower level is used
// If the policy is not set in any of these levels, a default value, defined in the policy package, is used
type InferenceProviderPolicies struct {
	Enabled *bool `toml:"enabled,omitempty"`
}

type InferencePolicies struct {
	// inference providers by property
	Property InferencePropertyPolicies `toml:"property,omitempty"`
	// inference providers by name
	Provider map[string]InferenceProviderPolicies `toml:"provider,omitempty"`
	// policies for all inference providers
	InferenceProviderPolicies
}

// Tools policies can be defined at different levels (highest priority first): by a provider name, or globally
// Policies are pointers, this means that if a policy is not set at a specific level, the value at a lower level is used
// If the policy is not set in any of these levels, a default value, defined in the policy package, is used
type ToolsProviderPolicies struct {
	Enabled        *bool `toml:"enabled,omitempty"`
	ReadOnly       *bool `toml:"read-only,omitempty"`
	NonDestructive *bool `toml:"non-destructive,omitempty"`
	// Local indicates if the tool cannot connect to a remote MCP server
	Local *bool `toml:"local,omitempty"`
}

type ToolsPolicies struct {
	// tools providers by property (tools do not have properties for the moment)

	// tools providers by name
	Provider map[string]ToolsProviderPolicies `toml:"provider,omitempty"`
	// policies for all tools providers
	ToolsProviderPolicies
}

type Policies struct {
	Inferences InferencePolicies `toml:"inferences,omitempty"`
	Tools      ToolsPolicies     `toml:"tools,omitempty"`
}

type PoliciesProvider interface {
	Read(policiesFile string) (*Policies, error)
	IsInferenceEnabledByPolicies(feature Feature[InferenceAttributes, InferenceInitializeOptions], policies *Policies) bool
	IsToolEnabledByPolicies(feature Feature[ToolsAttributes, ToolsInitializeOptions], policies *Policies) bool
	IsToolLocalByPolicies(feature Feature[ToolsAttributes, ToolsInitializeOptions], policies *Policies) bool
	IsToolNonDestructiveByPolicies(feature Feature[ToolsAttributes, ToolsInitializeOptions], policies *Policies) bool
	IsToolReadonlyByPolicies(feature Feature[ToolsAttributes, ToolsInitializeOptions], policies *Policies) bool
}
