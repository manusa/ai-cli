package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"

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
	commandAndArgs, err := getBestMcpServerCommand()
	if err != nil {
		return nil, err
	}
	cli, err := eino.StartMcp(ctx, commandAndArgs)
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Attributes())
}

func getBestMcpServerCommand() ([]string, error) {
	if commandExists("npx") {
		return []string{"npx", "-y", "kubernetes-mcp-server@latest"}, nil
	} else if commandExists("uvx") {
		return []string{"uvx", "kubernetes-mcp-server@latest"}, nil
	}
	// TODO support manual download and installation of kubernetes-mcp-server as a last resort
	return nil, errors.New("no command found to start the Kubernetes MCP server")
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
