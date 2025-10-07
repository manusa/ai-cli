package cmd

import (
	"os"
	"testing"

	"github.com/manusa/ai-cli/internal/test"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference/ollama"
	"github.com/manusa/ai-cli/pkg/keyring"
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
	keyring.MockInit()
	// Get the tmpdir before cleaning the environment
	// to avoid error on Windows
	tmpdir := s.T().TempDir()
	config.FileSystem = afero.NewBasePathFs(afero.NewOsFs(), tmpdir)

	s.originalEnv = os.Environ()
	os.Clearenv()

	ollama.DefaultBaseURL = "http://localhost:1337"

	s.rootCmd = NewAiCli()
}

func (s *DiscoverTestSuite) TearDownTest() {
	test.RestoreEnv(s.originalEnv)
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
			"    Description: Google Gemini inference provider\n" +
			"    Reason: GEMINI_API_KEY is not set\n" +
			"  - lmstudio\n" +
			"    Description: LM Studio local inference provider\n" +
			"    Reason: LM Studio is not accessible at http://localhost:1234\n" +
			"  - ollama\n" +
			"    Description: Ollama local inference provider\n" +
			"    Reason: ollama is not accessible at http://localhost:1337\n" +
			"  - ramalama\n" +
			"    Description: Ramalama local inference provider\n" +
			"    Reason: ramalama is not installed\n" +
			"Available Tools Providers:\n" +
			"  - fs\n" +
			"    Description: Provides access to the local filesystem, allowing listing of files and directories.\n" +
			"    Reason: filesystem is accessible\n" +
			"Not Available Tools Providers:\n" +
			"  - browsers\n" +
			"    Description: Provides access to browser metadata such as bookmarks, search history, and so on\n" +
			"    Reason: no browsers detected\n" +
			"  - github\n" +
			"    Description: Provides access to GitHub Platform. Provides the ability to to read repositories and code files, manage issues and PRs, analyze code, and automate workflows.\n" +
			"    Reason: GITHUB_PERSONAL_ACCESS_TOKEN is not set\n" +
			"  - kubernetes\n" +
			"    Description: Provides access to Kubernetes clusters, allowing management and interaction with cluster resources.\n" +
			"    Reason: no suitable MCP settings found for the Kubernetes MCP server\n" +
			"  - playwright\n" +
			"    Description: Enables web browsing capabilities through Playwright. Opening web pages, opening URLs, interacting with elements inside the browser, extracting snapshots, and scraping information from web pages. Support for multiple tabs and many other browser options\n" +
			"    Reason: npx command not found\n" +
			"  - postgresql\n" +
			"    Description: Provides access to a PostgreSQL database, allowing execution of SQL queries and retrieval of data.\n" +
			"    Reason: no suitable MCP settings found for the PostgreSQL MCP server\n"
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
			`{"description":"Google Gemini inference provider","name":"gemini","local":false,"public":true,"reason":"GEMINI_API_KEY is not set","models":null},` +
			`{"description":"LM Studio local inference provider","name":"lmstudio","local":true,"public":false,"reason":"LM Studio is not accessible at http://localhost:1234","models":null},` +
			`{"description":"Ollama local inference provider","name":"ollama","local":true,"public":false,"reason":"ollama is not accessible at http://localhost:1337","models":null},` +
			`{"description":"Ramalama local inference provider","name":"ramalama","local":true,"public":false,"reason":"ramalama is not installed","models":null}],` +
			`"inference":null,` +
			`"tools":[` +
			`{"description":"Provides access to the local filesystem, allowing listing of files and directories.","name":"fs","reason":"filesystem is accessible"}],` +
			`"toolsNotAvailable":[` +
			`{"description":"Provides access to browser metadata such as bookmarks, search history, and so on","name":"browsers","reason":"no browsers detected"},` +
			`{"description":"Provides access to GitHub Platform. Provides the ability to to read repositories and code files, manage issues and PRs, analyze code, and automate workflows.","name":"github","reason":"GITHUB_PERSONAL_ACCESS_TOKEN is not set"},` +
			`{"description":"Provides access to Kubernetes clusters, allowing management and interaction with cluster resources.","name":"kubernetes","reason":"no suitable MCP settings found for the Kubernetes MCP server"},` +
			`{"description":"Enables web browsing capabilities through Playwright. Opening web pages, opening URLs, interacting with elements inside the browser, extracting snapshots, and scraping information from web pages. Support for multiple tabs and many other browser options","name":"playwright","reason":"npx command not found"},` +
			`{"description":"Provides access to a PostgreSQL database, allowing execution of SQL queries and retrieval of data.","name":"postgresql","reason":"no suitable MCP settings found for the PostgreSQL MCP server"}` +
			`]}`
		s.JSONEq(expectedOutput, output, "Expected JSON output does not match")
	})
}

func (s *DiscoverTestSuite) TestMcpConfigUnknownEditor() {
	s.rootCmd.SetArgs([]string{"discover", "--mcp-config", "unknown"})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("Returns an error", func() {
		s.Empty(output, "Expected empty output")
		s.NotEmpty(err, "Expected error")
	})
}

func (s *DiscoverTestSuite) TestMcpConfigNoValue() {
	s.rootCmd.SetArgs([]string{"discover", "--mcp-config"})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("Returns an error", func() {
		s.Empty(output, "Expected empty output")
		s.NotEmpty(err, "Expected error")
	})
}

func (s *DiscoverTestSuite) TestMcpConfigEmptyValue() {
	s.rootCmd.SetArgs([]string{"discover", "--mcp-config", ""})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("Returns an error", func() {
		s.Empty(output, "Expected empty output")
		s.NotEmpty(err, "Expected error")
	})
}

func TestDiscover(t *testing.T) {
	suite.Run(t, new(DiscoverTestSuite))
}
