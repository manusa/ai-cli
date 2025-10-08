package kubernetes

import (
	"os"
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/keyring"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/stretchr/testify/suite"
)

type GithubTestSuite struct {
	suite.Suite
	originalEnv []string
}

func (s *GithubTestSuite) SetupTest() {
	keyring.MockInit()
	instance.Available = false
	s.originalEnv = os.Environ()
	os.Clearenv()
	inference.Clear()
	tools.Clear()
	tools.Register(instance)
}

func (s *GithubTestSuite) TearDownTest() {
	test.RestoreEnv(s.originalEnv)
}

func (s *GithubTestSuite) TestFeatureAttributes() {
	s.Run("feature name is github", func() {
		s.Equal("github", instance.FeatureName)
	})
	s.Run("has feature description", func() {
		s.Equal("Provides access to GitHub Platform. Provides the ability to to read repositories and code files, manage issues and PRs, analyze code, and automate workflows.", instance.FeatureDescription)
	})
}

func (s *GithubTestSuite) TestInitializeNoAccessToken() {
	feats := features.Discover(config.WithConfig(s.T().Context(), config.New()))
	s.Require().Empty(feats.Tools)
	s.Require().Len(feats.ToolsNotAvailable, 1)
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is not set, is not available", func() {
		s.False(feats.ToolsNotAvailable[0].IsAvailable())
	})
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is not set, shows reason", func() {
		s.Equal("GITHUB_PERSONAL_ACCESS_TOKEN is not set", feats.ToolsNotAvailable[0].Reason())
	})
}

func (s *GithubTestSuite) TestInitialize() {
	_ = os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "fake-token")
	feats := features.Discover(config.WithConfig(s.T().Context(), config.New()))
	s.Require().Len(feats.Tools, 1)
	s.Require().Empty(feats.ToolsNotAvailable)
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is set, is available", func() {
		s.True(feats.Tools[0].IsAvailable())
	})
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is set, shows reason", func() {
		s.Equal("GITHUB_PERSONAL_ACCESS_TOKEN is set", feats.Tools[0].Reason())
	})
	s.Run("sets MCP settings", func() {
		mcpSettings := feats.Tools[0].GetMcpSettings()
		s.Require().NotNil(mcpSettings, "McpSettings should be set")
		s.Run("sets Url", func() {
			s.Equal("https://api.githubcopilot.com/mcp/", mcpSettings.Url)
		})
		s.Run("sets Authorization header", func() {
			s.Contains(mcpSettings.Headers, "Authorization")
			s.Equal("Bearer fake-token", mcpSettings.Headers["Authorization"])
		})
		s.Run("sets only common toolsets in X-MCP-Toolsets header", func() {
			s.Contains(mcpSettings.Headers, "X-MCP-Toolsets")
			s.Equal("context,actions,issues,notifications,pull_requests,repos,users", mcpSettings.Headers["X-MCP-Toolsets"])
		})
		s.Run("does not set X-MCP-Readonly header by default", func() {
			s.NotContains(mcpSettings.Headers, "X-MCP-Readonly")
		})
	})
}

func (s *GithubTestSuite) TestInitializeReadOnly() {
	_ = os.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "fake-token")
	p := test.Must(policies.ReadToml(`
		[tools]
		read-only = true
	`))
	cfg := config.New()
	cfg.Enforce(p)
	feats := features.Discover(config.WithConfig(s.T().Context(), cfg))
	s.Require().Len(feats.Tools, 1)
	s.Require().Empty(feats.ToolsNotAvailable)
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is set, is available", func() {
		s.True(feats.Tools[0].IsAvailable())
	})
	s.Run("when GITHUB_PERSONAL_ACCESS_TOKEN is set, shows reason", func() {
		s.Equal("GITHUB_PERSONAL_ACCESS_TOKEN is set", feats.Tools[0].Reason())
	})
	s.Run("sets MCP settings", func() {
		mcpSettings := feats.Tools[0].GetMcpSettings()
		s.Require().NotNil(mcpSettings, "McpSettings should be set")
		s.Run("sets Url", func() {
			s.Equal("https://api.githubcopilot.com/mcp/", mcpSettings.Url)
		})
		s.Run("sets Authorization header", func() {
			s.Contains(mcpSettings.Headers, "Authorization")
			s.Equal("Bearer fake-token", mcpSettings.Headers["Authorization"])
		})
		s.Run("sets only common toolsets in X-MCP-Toolsets header", func() {
			s.Contains(mcpSettings.Headers, "X-MCP-Toolsets")
			s.Equal("context,actions,issues,notifications,pull_requests,repos,users", mcpSettings.Headers["X-MCP-Toolsets"])
		})
		s.Run("sets X-MCP-Readonly header when configured", func() {
			s.Contains(mcpSettings.Headers, "X-MCP-Readonly")
			s.Equal("true", mcpSettings.Headers["X-MCP-Readonly"])
		})
	})
}

func TestGithub(t *testing.T) {
	suite.Run(t, new(GithubTestSuite))
}
