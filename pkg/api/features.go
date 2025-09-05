package api

import (
	"cmp"
	"context"
)

type Feature[a FeatureAttributes] interface {
	Attributes() a
	// Initialize Performs the discovery and initialization of the feature based on the user configuration and policies
	// Populates the internal state of the feature and its availability
	Initialize(ctx context.Context)
	// IsAvailable Returns true if the feature is available
	IsAvailable() bool
	// Reason provides the reason why the feature is or is not available
	Reason() string
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

type IsFeatureEnabled[a FeatureAttributes] func(feature Feature[a]) bool

func FeatureSorter[A FeatureAttributes, F Feature[A]](a F, b F) int {
	return cmp.Compare(a.Attributes().Name(), b.Attributes().Name())
}
