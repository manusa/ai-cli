package api

import "github.com/manusa/ai-cli/pkg/config"

type Feature[a FeatureAttributes] interface {
	Attributes() a
	IsAvailable(cfg *config.Config) bool
}

type FeatureAttributes interface {
	Name() string
}

type BasicFeatureAttributes struct {
	FeatureName string
}

func (b BasicFeatureAttributes) Name() string {
	return b.FeatureName
}
