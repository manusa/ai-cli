package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/yaml"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/spf13/cobra"
)

type DiscoverCmdOptions struct {
	outputFormat   string
	policiesFile   string
	policiesSample bool
}

func NewDiscoverCmdOptions() *DiscoverCmdOptions {
	return &DiscoverCmdOptions{}
}

// NewDiscoverCmd creates a new command to discover AI capabilities for the current system
// TODO: rename to "capabilities" or "features"?
func NewDiscoverCmd() *cobra.Command {
	o := NewDiscoverCmdOptions()
	cmd := &cobra.Command{
		Use:   "discover", // TODO: rename to "capabilities" or "features"?
		Short: "Discover AI capabilities for the current system",
		Long:  "Discover available AI capabilities (llm providers, models, applicable tools) for the current system",
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

	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", "json", "Output format (json, text)")
	cmd.Flags().StringVar(&o.policiesFile, "policies", "", "Policies file to use")
	cmd.Flags().BoolVar(&o.policiesSample, "show-policies-sample", false, "Outputs sample policies file")

	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *DiscoverCmdOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *DiscoverCmdOptions) Validate() error {
	// TODO: validate output format
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *DiscoverCmdOptions) Run(_ *cobra.Command) error {
	if o.policiesSample {
		policies := features.GetDefaultPolicies()
		bytes, err := yaml.Marshal(policies)
		if err != nil {
			return fmt.Errorf("failed to marshal default policies to YAML: %w", err)
		}
		_, _ = fmt.Printf("%s\n", bytes)
		return nil
	}

	var userPolicies *policies.Policies
	if len(o.policiesFile) > 0 {
		var err error
		userPolicies, err = policies.Read(o.policiesFile)
		if err != nil {
			return fmt.Errorf("failed to read preferences: %w", err)
		}
	}

	discoveredFeatures := features.Discover(config.New(), userPolicies)
	// TODO: maybe create an output package to handle different output formats globally
	switch o.outputFormat {
	case "json":
		bytes, err := json.Marshal(discoveredFeatures)
		if err != nil {
			return fmt.Errorf("failed to marshal discovered features to JSON: %w", err)
		}
		_, _ = fmt.Printf("%s\n", bytes)
	case "text":
		_, _ = fmt.Print(discoveredFeatures.ToHumanReadable())
	}
	return nil
}
