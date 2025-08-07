package api

import "github.com/manusa/ai-cli/pkg/config"

type BasicFeatureProvider struct {
	Reason string
}

type Feature[a FeatureAttributes, b any] interface {
	Attributes() a
	Data() b
	IsAvailable(cfg *config.Config) bool
}

type FeatureAttributes interface {
	Name() string
}

type BasicFeatureAttributes struct {
	FeatureName string `json:"name"`
}

func (b BasicFeatureAttributes) Name() string {
	return b.FeatureName
}

type BasicFeatureData struct {
	Reason string `json:"reason"`
}
