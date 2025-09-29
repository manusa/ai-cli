package kubernetes

import (
	"context"
	"fmt"
	"os"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/tools"
)

type Provider struct {
	api.BasicToolsProvider
}

var _ api.ToolsProvider = &Provider{}

const (
	accessTokenEnvVar = "GITHUB_PERSONAL_ACCESS_TOKEN"
)

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.ToolsParameters = cfg.ToolsParameters(p.Attributes().Name())
	}

	accessToken := os.Getenv(accessTokenEnvVar)
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

var instance = &Provider{
	BasicToolsProvider: api.BasicToolsProvider{
		BasicToolsAttributes: api.BasicToolsAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "github",
				FeatureDescription: "Provides access to GitHub Platform. Provides the ability to to read repositories and code files, manage issues and PRs, analyze code, and automate workflows.",
			},
		},
	},
}

func init() {
	tools.Register(instance)
}
