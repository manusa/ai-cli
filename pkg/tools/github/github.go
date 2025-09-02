package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	api.BasicToolsProvider
	ReadOnly bool `json:"-"`
}

var _ api.ToolsProvider = &Provider{}

type GithubPolicies struct {
	policies.ToolPolicies
}

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

func (p *Provider) Initialize(_ context.Context, toolPolicies any) {
	// TODO: This should probably be generalized to all tools and inference providers
	if !policies.IsEnabledByPolicies(toolPolicies) {
		p.IsAvailableReason = "github is not authorized by policies"
		return
	}

	if policies.IsReadOnlyByPolicies(toolPolicies) {
		p.ReadOnly = true
	}

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

func (p *Provider) GetDefaultPolicies() map[string]any {
	var policies = GithubPolicies{}
	jsonBody, err := json.Marshal(policies)
	if err != nil {
		return nil
	}
	var policiesMap map[string]any
	err = json.Unmarshal(jsonBody, &policiesMap)
	if err != nil {
		return nil
	}
	return policiesMap
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
