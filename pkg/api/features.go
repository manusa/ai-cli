package api

import "github.com/manusa/ai-cli/pkg/config"

type Feature[a FeatureAttributes] interface {
	Attributes() a
	// IsAvailable Returns true if the feature is available based on the user configuration and policies
	IsAvailable(cfg *config.Config, policies any) bool
	// Reason provides the reason why the feature is or is not available
	Reason() string
	GetDefaultPolicies() map[string]any
}

type FeatureAttributes interface {
	// Name of the feature
	Name() string
	// Description of the feature
	Description() string
}

type BasicFeatureAttributes struct {
	FeatureAttributes  `json:"-"`
	FeatureName        string `json:"name"`
	FeatureDescription string `json:"description"`
}

func (a *BasicFeatureAttributes) Name() string {
	return a.FeatureName
}

func (a *BasicFeatureAttributes) Description() string {
	return a.FeatureDescription
}
