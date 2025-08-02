package kubernetes

import (
	"context"
	"encoding/json"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
	"k8s.io/client-go/tools/clientcmd"
)

type Provider struct {
}

var _ tools.Provider = &Provider{}

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "kubernetes",
		},
	}
}

func (p *Provider) IsAvailable(_ *config.Config) bool {
	_, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	return err == nil
}

func (p *Provider) GetTools(ctx context.Context, _ *config.Config) ([]*api.Tool, error) {
	// TODO: start the MCP Server
	cli, err := eino.StartMcpClient(ctx)
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Attributes())
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
