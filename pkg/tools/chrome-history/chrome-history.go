package chrome

import (
	"context"
	"encoding/json"

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
			FeatureName: "chrome-history",
		},
		ModelAttributes: &api.ModelAttributes{
			NeededTools: []string{"fs", "os"},
			Prompt: `If my system is Windows, is the file 'History' exists in the directory 'AppData/Local/Google/Chrome/UserData/Default' in the home directory?
Or if my system is Darwin, is the file 'History' exists in the directory 'Library/Application Support/Google/Chrome/Default' in the home directory?
Or if the system is Linux, is the file 'History' exists in the directory '.config/google-chrome/Default' in the home directory?
You can use tools without asking confirmation. If Yes, you must reply with Yes only. If No, explain why`,
		},
	}
}

func (p *Provider) IsAvailable(_ *config.Config) bool {
	// not use, because ModelAttributes takes precedence
	return false
}

func (p *Provider) Data() tools.Data {
	return tools.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: p.Reason,
		},
	}
}

func (p *Provider) GetTools(_ context.Context, _ *config.Config) ([]*api.Tool, error) {
	return []*api.Tool{}, nil
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(tools.Report{
		Attributes: p.Attributes(),
		Data:       p.Data(),
	})
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
