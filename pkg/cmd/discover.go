package cmd

import (
	"fmt"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/feature"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/spf13/cobra"
)

type DiscoverCmdOptions struct{}

func NewDiscoverCmdOptions() *DiscoverCmdOptions {
	return &DiscoverCmdOptions{}
}

func NewDiscoverCmd() *cobra.Command {
	o := NewDiscoverCmdOptions()
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover features",
		Long:  "Discover features",
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
func (o *DiscoverCmdOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *DiscoverCmdOptions) Validate() error {
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *DiscoverCmdOptions) Run(cmd *cobra.Command) error {
	cfg := config.New() // TODO, will need to infer or load from a file

	// very simple for now, just print the available inferences and features
	if err := o.discoverInferences(cfg); err != nil {
		return err
	}
	if err := o.discoverFeatures(cfg); err != nil {
		return err
	}

	return nil
}

func (o *DiscoverCmdOptions) discoverInferences(cfg *config.Config) error {
	inferences, err := inference.GetAvailableInferences(cfg)
	if err != nil {
		return fmt.Errorf("failed to get available inferences: %w", err)
	}
	fmt.Printf("available inferences:\n")
	for _, i := range inferences {
		fmt.Printf("  - %s\n", i.Name)
	}
	fmt.Println()
	return nil
}

func (o *DiscoverCmdOptions) discoverFeatures(cfg *config.Config) error {
	features, err := feature.GetAvailableFeatures(cfg)
	if err != nil {
		return fmt.Errorf("failed to get available features: %w", err)
	}

	fmt.Printf("available features:\n")
	for _, f := range features {
		fmt.Printf("  - %s\n", f.Name)
	}
	fmt.Println()
	return nil
}
