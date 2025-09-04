package playwright

import (
	"os"
	"strings"
	"testing"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/stretchr/testify/suite"
)

type PlaywrightTestSuite struct {
	suite.Suite
	originalEnv []string
}

func (s *PlaywrightTestSuite) SetupTest() {
	s.originalEnv = os.Environ()
	os.Clearenv()
	inference.Clear()
	tools.Clear()
	tools.Register(instance)
}

func (s *PlaywrightTestSuite) TearDownTest() {
	os.Clearenv()
	for _, env := range s.originalEnv {
		if key, value, found := strings.Cut(env, "="); found {
			_ = os.Setenv(key, value)
		}
	}
}

func (s *PlaywrightTestSuite) TestFeatureAttributes() {
	s.Run("feature name is playwright", func() {
		s.Equal("playwright", instance.FeatureName)
	})
	s.Run("has feature description", func() {
		s.Equal("Automate and interact with web browsers using Playwright.", instance.FeatureDescription)
	})
}

func (s *PlaywrightTestSuite) TestInitializeWithNoNode() {
	feats := features.Discover(config.WithConfig(s.T().Context(), config.New()))
	s.Require().Empty(feats.Tools)
	s.Require().Len(feats.ToolsNotAvailable, 1)
	s.Run("when npx command does not exist, is not available", func() {
		s.False(feats.ToolsNotAvailable[0].IsAvailable())
	})
	s.Run("when npx command does not exist, shows reason", func() {
		s.Equal(feats.ToolsNotAvailable[0].Reason(), "npx command not found")
	})
}

func (s *PlaywrightTestSuite) TestInitializeWithNodeAndDesktop() {
	config.LookPath = func(path string) (string, error) { return "/path/to/npx", nil }
	_ = os.Setenv("DISPLAY", ":0")
	feats := features.Discover(config.WithConfig(s.T().Context(), config.New()))
	s.Require().Len(feats.Tools, 1)
	s.Require().Empty(feats.ToolsNotAvailable)
	s.Run("when npx command exists, is available", func() {
		s.True(feats.Tools[0].IsAvailable())
	})
	s.Run("when npx command exists, shows reason", func() {
		s.Equal(feats.Tools[0].Reason(), "npx command found")
	})
	s.Run("sets MCP settings", func() {
		s.Equal("npx", feats.Tools[0].(*Provider).McpSettings.Command)
		s.Equal([]string{"-y", "@playwright/mcp@0.0.36"}, feats.Tools[0].(*Provider).McpSettings.Args)
	})
}

func (s *PlaywrightTestSuite) TestInitializeWithNodeAndNoDesktop() {
	config.LookPath = func(path string) (string, error) { return "/path/to/npx", nil }
	config.Os = "linux"
	feats := features.Discover(config.WithConfig(s.T().Context(), config.New()))
	s.Require().Len(feats.Tools, 1)
	s.Require().Empty(feats.ToolsNotAvailable)
	s.Run("sets MCP settings with headless", func() {
		s.Equal("npx", feats.Tools[0].(*Provider).McpSettings.Command)
		s.Equal([]string{"-y", "@playwright/mcp@0.0.36", "--headless"}, feats.Tools[0].(*Provider).McpSettings.Args)
	})
}

func TestPlaywright(t *testing.T) {
	suite.Run(t, new(PlaywrightTestSuite))
}
