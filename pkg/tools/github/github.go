package kubernetes

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/keyring"
	"github.com/manusa/ai-cli/pkg/system"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/manusa/ai-cli/pkg/ui/components/password_input"
	"github.com/manusa/ai-cli/pkg/ui/components/selector"
)

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

const (
	accessTokenEnvVar               = "GITHUB_PERSONAL_ACCESS_TOKEN"
	createNewPersonalAccessTokenUrl = "https://github.com/settings/personal-access-tokens/new"
)

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.ToolsParameters = cfg.ToolsParameters(p.Attributes().Name())
	}

	accessToken := p.getAccessToken()
	if accessToken == "" {
		p.IsAvailableReason = fmt.Sprintf("%s is not set", accessTokenEnvVar)
		return
	}

	headers := map[string]string{
		"Authorization":  "Bearer " + accessToken,
		"X-MCP-Toolsets": "context,actions,issues,notifications,pull_requests,repos,users",
	}
	if *p.ReadOnly {
		headers["X-MCP-Readonly"] = "true"
	}
	p.IsAvailableReason = fmt.Sprintf("%s is set", accessTokenEnvVar)
	p.Available = true
	p.McpSettings = &api.McpSettings{
		Type:    api.McpTypeStreamableHttp,
		Url:     "https://api.githubcopilot.com/mcp/",
		Headers: headers,
	}
}

func (p *Provider) getAccessToken() string {
	if key, err := keyring.GetKey(accessTokenEnvVar); err == nil && len(key) > 0 {
		return key
	}
	return os.Getenv(accessTokenEnvVar)
}

func (p *Provider) InstallHelp() error {
	createNewPersonalAccessToken := "Create a new GitHub Personal Access Token"
	registerExistingPersonalAccessToken := "Register an existing GitHub Personal Access Token"
	quit := "Terminate GitHub setup"
	choices := []list.Item{
		selector.Item(createNewPersonalAccessToken),
		selector.Item(registerExistingPersonalAccessToken),
		selector.Item(quit),
	}
	for {
		choice, err := selector.Select("Please select a step:", choices)
		if err != nil {
			return err
		}
		switch choice {
		case createNewPersonalAccessToken:
			fmt.Printf("Opening browser to create a new personal access token...\nYou can also access the page at %s\n", createNewPersonalAccessTokenUrl)
			err = system.OpenBrowser(createNewPersonalAccessTokenUrl)
			if err != nil {
				return err
			}
		case registerExistingPersonalAccessToken:
			fmt.Printf("Paste your token below:\n")
			apiKey, err := password_input.Prompt()
			if err != nil {
				return err
			}
			err = keyring.SetKey(accessTokenEnvVar, apiKey)
			if err != nil {
				return err
			}
		case quit:
			return nil
		}
	}
}

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "github",
				FeatureDescription: "Provides access to GitHub Platform. Provides the ability to to read repositories and code files, manage issues and PRs, analyze code, and automate workflows.",
				SupportsSetupAttr:  true,
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
