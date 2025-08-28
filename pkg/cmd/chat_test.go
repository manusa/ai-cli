package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

type ChatTestSuite struct {
	suite.Suite
	rootCmd *cobra.Command
}

func (s *ChatTestSuite) SetupTest() {
	s.rootCmd = NewAiCli()
}

func (s *ChatTestSuite) TestHelp() {
	s.rootCmd.SetArgs([]string{"chat", "--help"})
	output, err := captureOutput(s.rootCmd.Execute)
	s.Run("--help command exists", func() {
		s.Nilf(err, "Expected no error, got: %v", err)
	})
	s.Run("Shows long description message", func() {
		s.Contains(output, "Start an interactive chat with an AI model\n")
	})
	s.Run("Shows usage message", func() {
		s.Contains(output, ""+
			"Usage:\n"+
			"  ai-cli chat [flags]\n"+
			"\n")
	})
	s.Run("No flags are advertised", func() {
		s.Regexp(""+
			"Flags:\n"+
			"  -h, --help   help for chat\n$", output)
	})
}

func TestChat(t *testing.T) {
	suite.Run(t, new(ChatTestSuite))
}
