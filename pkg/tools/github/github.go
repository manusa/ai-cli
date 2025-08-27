package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/tools/utils/eino"
)

type Provider struct {
	tools.BasicToolsProvider
	ReadOnly bool
}

type GithubPolicies struct {
	policies.ToolPolicies
}

const (
	accessTokenEnvVar = "GITHUB_PERSONAL_ACCESS_TOKEN"
)

var _ tools.Provider = &Provider{}

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

func (p *Provider) Attributes() tools.Attributes {
	return tools.Attributes{
		BasicFeatureAttributes: api.BasicFeatureAttributes{
			FeatureName: "github",
		},
	}
}

func (p *Provider) Data() tools.Data {
	data := tools.Data{
		BasicFeatureData: api.BasicFeatureData{
			Reason: p.Reason,
		},
	}
	settings, err := findBestMcpServerSettings(p.ReadOnly)
	if err == nil {
		data.McpSettings = settings
	}
	return data
}

func (p *Provider) IsAvailable(_ *config.Config, toolPolicies any) bool {
	if !policies.IsEnabledByPolicies(toolPolicies) {
		p.Reason = "github is not authorized by policies"
		return false
	}

	if policies.IsReadOnlyByPolicies(toolPolicies) {
		p.ReadOnly = true
	}

	available := os.Getenv(accessTokenEnvVar) != ""
	if available {
		p.Reason = fmt.Sprintf("%s is set", accessTokenEnvVar)
	} else {
		p.Reason = fmt.Sprintf("%s is not set", accessTokenEnvVar)
	}
	return available
}

func (p *Provider) GetTools(ctx context.Context, _ *config.Config) ([]*api.Tool, error) {
	mcpSettings, err := findBestMcpServerSettings(p.ReadOnly)
	if err != nil || mcpSettings.Type != api.McpTypeStdio {
		return nil, err
	}

	cli, err := eino.StartMcp(ctx, slices.Concat([]string{mcpSettings.Command}, mcpSettings.Args))
	if err != nil {
		return nil, err
	}
	return eino.GetTools(ctx, cli)
}

func (p *Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(tools.Report{
		Attributes: p.Attributes(),
		Data:       p.Data(),
	})
}

func findBestMcpServerSettings(readOnly bool) (*api.McpSettings, error) {
	for command, settings := range supportedMcpSettings {
		if commandExists(command) {
			if readOnly {
				settings.Args = append(settings.Args, "--read-only")
			}
			return &settings, nil
		}
	}
	return nil, errors.New("no suitable MCP settings found for the Github MCP server")
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
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

var instance = &Provider{}

func init() {
	tools.Register(instance)
}
