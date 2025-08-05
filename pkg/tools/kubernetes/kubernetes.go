package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"syscall"

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
	err := startKubernetesMcpServer(ctx)
	if err != nil {
		return nil, err
	}
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

func startKubernetesMcpServer(ctx context.Context) error {
	command, err := getBestCommand()
	if err != nil {
		return fmt.Errorf("failed to get best command: %v", err)
	}
	if len(command) == 0 {
		return fmt.Errorf("command must have at least one element")
	}
	port, err := GetFreePort()
	if err != nil {
		return fmt.Errorf("failed to get free port: %v", err)
	}
	args := command[1:]
	args = append(args, "--port", strconv.Itoa(port))
	cmd := exec.CommandContext(ctx, command[0], args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	// TODO redirect stdout and stderr somewhere https://github.com/manusa/ai-cli/issues/19

	go func() {
		err = cmd.Run()
		if err != nil {
			log.Printf("Kubernetes MCP server terminates: %v", err)
		}
	}()
	return nil
}

func getBestCommand() ([]string, error) {
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

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer func() { _ = l.Close() }()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
