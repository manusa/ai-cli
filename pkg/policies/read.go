package policies

import (
	"os"

	"github.com/invopop/yaml"
)

func Read(policiesFile string) (*Policies, error) {
	policies := Policies{}
	fileContent, err := os.ReadFile(policiesFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(fileContent, &policies)
	if err != nil {
		return nil, err
	}
	return &policies, nil
}
