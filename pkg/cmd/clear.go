package cmd

import (
	"fmt"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/setup"
	"github.com/spf13/cobra"
)

type ClearCmdOptions struct {
	policiesFile string

	Logger
}

func NewClearCmdOptions() *ClearCmdOptions {
	return &ClearCmdOptions{}
}

// NewClearCmd creates a new command to clear the secrets stored during setup
func NewClearCmd() *cobra.Command {
	o := NewClearCmdOptions()
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the secrets stored during setup",
		Long:  "Help the user to clear the secrets stored during setup",
		RunE: func(cmd *cobra.Command, args []string) error {

			o.initLogger()
			defer o.disposeLogger()

			// Reuse k8s cli complete,validate,run pattern: https://github.com/kubernetes/sample-cli-plugin/blob/7922d71292adb0b472d54d7e03e8daa6eeb46576/pkg/cmd/ns.go
			if err := o.Complete(cmd, args); err != nil {
				return err
			}
			if err := o.Validate(cmd); err != nil {
				return err
			}
			if err := o.Run(cmd); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&o.policiesFile, "policies", "", "Policies file to use")

	o.initLoggerFlags(cmd)

	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *ClearCmdOptions) Complete(cmd *cobra.Command, _ []string) error {
	cfg := config.New()

	var userPolicies *api.Policies
	if len(o.policiesFile) > 0 {
		var err error
		userPolicies, err = policies.PoliciesProvider.Read(o.policiesFile)
		if err != nil {
			return fmt.Errorf("failed to read preferences: %w", err)
		}
	}
	cfg.Enforce(userPolicies)
	cmd.SetContext(config.WithConfig(cmd.Context(), cfg))

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *ClearCmdOptions) Validate(cmd *cobra.Command) error {
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *ClearCmdOptions) Run(cmd *cobra.Command) error {
	return setup.ClearAll(cmd.Context())
}
