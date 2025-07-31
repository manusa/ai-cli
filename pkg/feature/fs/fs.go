package fs

import (
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/feature"
)

type FSProvider struct{}

var fsProvider = FSProvider{}

func init() {
	feature.Register(fsProvider)
}

func (o FSProvider) Attributes() feature.FeatureAttributes {
	return feature.FeatureAttributes{
		Name: "fs",
	}
}

func (o FSProvider) IsAvailable(cfg *config.Config) bool {
	return true
}
