package os

import (
	"context"
	"encoding/json"
	"runtime"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Provider struct {
	tools.BasicToolsProvider
}

var _ tools.Provider = &Provider{}

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "os",
		},
	}
}

func (p *Provider) IsAvailable(_ *config.Config) bool {
	p.Reason = "os is accessible"
	return true
}

func (p *Provider) Data() tools.Data {
	return tools.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: p.Reason,
		},
	}
}

func (p *Provider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return []*api.Tool{
		GetOS,
	}, nil
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(tools.Report{
		Attributes: p.Attributes(),
		Data:       p.Data(),
	})
}

var GetOS = &api.Tool{
	Name: "get_os",
	Description: "Returns the name of the operating system (OS) of the user." +
		"Returns the name of the OS.",
	Parameters: map[string]api.ToolParameter{},
	Function: func(args map[string]interface{}) (string, error) {
		return runtime.GOOS, nil
	},
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
