package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	api.BasicToolsProvider
	ReadOnly bool `json:"-"`
}

var _ api.ToolsProvider = &Provider{}

const (
	accessTokenEnvVar = "GITHUB_PERSONAL_ACCESS_TOKEN"
)

var (
	supportedMcpSettings = map[string]api.McpSettings{
		"podman": {
			Type:    api.McpTypeStdio,
			Command: "podman",
			Args: []string{
				"run",
				"-i",
				"--rm",
				"-e",
				accessTokenEnvVar,
				"--entrypoint",
				"/server/github-mcp-server",
				"ghcr.io/github/github-mcp-server",
				"stdio",
			},
		},
	}
)

func (p *Provider) Initialize(_ context.Context) {
	hasAccessToken := os.Getenv(accessTokenEnvVar) != ""
	if !hasAccessToken {
		p.IsAvailableReason = fmt.Sprintf("%s is not set", accessTokenEnvVar)
		return
	}

	var err error
	p.McpSettings, err = findBestMcpServerSettings(p.ReadOnly)
	if err != nil {
		p.IsAvailableReason = err.Error()
		return
	}

	p.Available = true
	p.IsAvailableReason = fmt.Sprintf("%s is set and has suitable MCP settings", accessTokenEnvVar)
}

func (p *Provider) GetTools(ctx context.Context) ([]*api.Tool, error) {
	mcpSettings, err := findBestMcpServerSettings(p.ReadOnly)
	if err != nil || mcpSettings.Type != api.McpTypeStdio {
		return nil, err
	}

	cli, err := eino.StartMcp(ctx, mcpSettings.Env, slices.Concat([]string{mcpSettings.Command}, mcpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func findBestMcpServerSettings(readOnly bool) (*api.McpSettings, error) {
	for command, settings := range supportedMcpSettings {
		if config.CommandExists(command) {
			if readOnly {
				settings.Args = append(settings.Args, "--read-only")
			}
			return &settings, nil
		}
	}
	return nil, errors.New("no suitable MCP settings found for the Github MCP server")
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "github",
				FeatureDescription: "Provides access to GitHub repositories, issues, pull requests, and more.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
