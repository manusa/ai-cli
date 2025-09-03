package api

type InferencePropertyPolicies struct {
	Remote InferenceProviderPolicies `toml:"remote,omitempty"`
	Public InferenceProviderPolicies `toml:"public,omitempty"`
}

// InferenceProviderPolicies struct to define policies for inference providers
//
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

// ToolsProviderPolicies struct to define policies for tools providers
//
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

type PolicyVerifier[a FeatureAttributes] func(feature Feature[a], policies *Policies) (result bool, enforced bool)

type PoliciesProvider interface {
	Read(policiesFile string) (*Policies, error)
	IsInferenceEnabledByPolicies(feature Feature[InferenceAttributes], policies *Policies) (result bool, enforced bool)
	IsToolEnabledByPolicies(feature Feature[ToolsAttributes], policies *Policies) (result bool, enforced bool)
	IsToolLocalByPolicies(feature Feature[ToolsAttributes], policies *Policies) (result bool, enforced bool)
	IsToolNonDestructiveByPolicies(feature Feature[ToolsAttributes], policies *Policies) (result bool, enforced bool)
	IsToolReadonlyByPolicies(feature Feature[ToolsAttributes], policies *Policies) (result bool, enforced bool)
}
