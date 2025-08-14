package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

type DiscoverTestSuite struct {
	suite.Suite
	rootCmd     *cobra.Command
	originalEnv []string
}

func (s *DiscoverTestSuite) SetupTest() {
	s.originalEnv = os.Environ()
	os.Clearenv()
	config.FileSystem = afero.NewBasePathFs(afero.NewOsFs(), s.T().TempDir())
	s.rootCmd = NewAiCli()
}

func (s *DiscoverTestSuite) TearDownTest() {
	os.Clearenv()
	for _, env := range s.originalEnv {
		if key, value, found := strings.Cut(env, "="); found {
			_ = os.Setenv(key, value)
		}
	}
}

func (s *DiscoverTestSuite) TestOutputText() {
	s.rootCmd.SetArgs([]string{"discover", "--output", "text"})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("Returns no error", func() {
		s.NotEmpty(output, "Expected non-empty output")
		s.NoErrorf(err, "Error executing command: %v", err)
	})
	s.Run("Outputs human-readable text", func() {
		expectedOutput := "Available Inference Providers:\n" +
			"Not Available Inference Providers:\n" +
			"  - gemini\n" +
			"    Reason: GEMINI_API_KEY is not set\n" +
			"  - ollama\n" +
			"    Reason: http://localhost:11434 is not accessible\n" +
			"Available Tools Providers:\n" +
			"  - fs\n" +
			"    Reason: filesystem is accessible\n" +
			"Not Available Tools Providers:\n" +
			"  - kubernetes\n" +
			"    Reason: no kubeconfig file found in the default location\n"
		s.Equal(expectedOutput, output, "Expected output does not match")
	})
}

func (s *DiscoverTestSuite) TestOutputJson() {
	s.rootCmd.SetArgs([]string{"discover", "--output", "json"})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("Returns no error", func() {
		s.NotEmpty(output, "Expected non-empty output")
		s.NoErrorf(err, "Error executing command: %v", err)
	})
	s.Run("Outputs valid JSON", func() {
		expectedOutput := "{" +
			`"inferences":[],` +
			`"inferencesNotAvailable":[` +
			`{"name":"gemini","local":false,"public":true,"reason":"GEMINI_API_KEY is not set","models":null},` +
			`{"name":"ollama","local":true,"public":false,"reason":"http://localhost:11434 is not accessible","models":null}],` +
			`"inference":null,` +
			`"tools":[{"name":"fs","reason":"filesystem is accessible"}],` +
			`"toolsNotAvailable":[{"name":"kubernetes","reason":"no kubeconfig file found in the default location"}]}`
		s.JSONEq(expectedOutput, output, "Expected JSON output does not match")
	})
}

func TestDiscoverText(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
