package cmd

import (
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/spf13/cobra"
)

func NewAiCli() *cobra.Command {
	cmd := &cobra.Command{
		Use:   version.BinaryName,
		Short: "AI CLI",
		Long:  "AI CLI is a command line interface for interacting with AI models and services.",
	}

	cmd.AddCommand(NewChatCmd())
	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewSetupCmd())
	cmd.AddCommand(NewVersionCmd())

	return cmd
}
