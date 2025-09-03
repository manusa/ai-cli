package policies

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/manusa/ai-cli/pkg/api"
)

func (p *Provider) Read(policiesFile string) (*api.Policies, error) {
	fileContent, err := os.ReadFile(policiesFile)
	if err != nil {
		return nil, err
	}
	return ReadToml(string(fileContent))
}

func ReadToml(config string) (*api.Policies, error) {
	policies := api.Policies{}
	_, err := toml.Decode(config, &policies)
	if err != nil {
		return nil, err
	}
	return &policies, nil
}
