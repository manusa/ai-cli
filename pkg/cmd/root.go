package cmd

import (
	"fmt"
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/spf13/cobra"
)

type AiCliOptions struct {
	Version bool
}

func NewAiCliOptions() *AiCliOptions {
	return &AiCliOptions{}
}

func NewAiCli() *cobra.Command {
	o := NewAiCliOptions()
	cmd := &cobra.Command{
		Use:   version.BinaryName,
		Short: "AI CLI",
		Long:  "AI CLI is a command line interface for interacting with AI models and services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Reuse k8s cli complete,validate,run pattern: https://github.com/kubernetes/sample-cli-plugin/blob/7922d71292adb0b472d54d7e03e8daa6eeb46576/pkg/cmd/ns.go
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&o.Version, "version", false, "Print version information and quit")

	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *AiCliOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *AiCliOptions) Validate() error {
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *AiCliOptions) Run() error {

	if o.Version {
		_, _ = fmt.Printf("%s\n", version.Version)
		return nil
	}

	return nil
}
