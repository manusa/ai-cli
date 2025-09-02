package api

type InferencePropertyPolicies struct {
	Remote InferenceProviderPolicies `toml:"remote,omitempty"`
	Public InferenceProviderPolicies `toml:"public,omitempty"`
}

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
	IsInferenceEnabledByPolicies(feature Feature[InferenceAttributes], policies *Policies) bool
	IsToolEnabledByPolicies(feature Feature[ToolsAttributes], policies *Policies) bool
	IsToolLocalByPolicies(feature Feature[ToolsAttributes], policies *Policies) bool
	IsToolNonDestructiveByPolicies(feature Feature[ToolsAttributes], policies *Policies) bool
	IsToolReadonlyByPolicies(feature Feature[ToolsAttributes], policies *Policies) bool
}
