package policies

import (
	"encoding/json"
)

type Policies struct {
	Tools map[string]any `yaml:"tools"`
}

type ToolPolicies struct {
	Enabled  bool `yaml:"enabled" json:"enabled"`
	ReadOnly bool `yaml:"read-only" json:"read-only"`
}

// IsEnabledByPolicies checks if the tool is enabled by policies
// If the tool policies are nil, it returns true
// If the tool policies are not nil, it returns the value of the Enabled field
// If the tool policies are not a valid ToolPolicies struct, it returns false
func IsEnabledByPolicies(toolPolicies any) bool {
	if toolPolicies == nil {
		return true
	}
	jsonBody, err := json.Marshal(toolPolicies)
	if err != nil {
		return false
	}
	var structuredPolicies ToolPolicies
	err = json.Unmarshal(jsonBody, &structuredPolicies)
	if err != nil {
		return false
	}
	return structuredPolicies.Enabled
}

// IsReadOnlyByPolicies checks if the tool must be read-only by policies
// If the tool policies are nil, it returns false
// If the tool policies are not nil, it returns the value of the ReadOnly field
// If the tool policies are not a valid ToolPolicies struct, it returns true
func IsReadOnlyByPolicies(toolPolicies any) bool {
	if toolPolicies == nil {
		return false
	}
	jsonBody, err := json.Marshal(toolPolicies)
	if err != nil {
		return true
	}
	var structuredPolicies ToolPolicies
	err = json.Unmarshal(jsonBody, &structuredPolicies)
	if err != nil {
		return true
	}
	return structuredPolicies.ReadOnly
}
