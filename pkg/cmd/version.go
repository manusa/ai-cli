package cmd

import (
	"fmt"

	"github.com/manusa/ai-cli/pkg/version"
	"github.com/spf13/cobra"
)

type VersionCmdOptions struct{}

func NewVersionCmdOptions() *VersionCmdOptions {
	return &VersionCmdOptions{}
}

func NewVersionCmd() *cobra.Command {
	o := NewVersionCmdOptions()
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  "Show version of the CLI and quit",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Reuse k8s cli complete,validate,run pattern: https://github.com/kubernetes/sample-cli-plugin/blob/7922d71292adb0b472d54d7e03e8daa6eeb46576/pkg/cmd/ns.go
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(cmd); err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *VersionCmdOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *VersionCmdOptions) Validate() error {
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *VersionCmdOptions) Run(cmd *cobra.Command) error {
	_, _ = fmt.Printf("%s\n", version.Version)

	return nil
}
