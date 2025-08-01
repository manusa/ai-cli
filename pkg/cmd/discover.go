package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/spf13/cobra"
)

type DiscoverCmdOptions struct {
	outputFormat string
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
	discoveredFeatures := features.Discover(config.New())
	// TODO: maybe create an output package to handle different output formats globally
	switch o.outputFormat {
	case "json":
		bytes, err := json.Marshal(discoveredFeatures)
		if err != nil {
			return fmt.Errorf("failed to marshal discovered features to JSON: %w", err)
		}
		_, _ = fmt.Printf("%s\n", bytes)
	case "text":
		_, _ = fmt.Printf("Available Inference Providers:\n")
		for _, provider := range discoveredFeatures.Inferences {
			fmt.Printf("  - %s\n", provider.Attributes().Name())
			models, err := provider.GetModels(context.Background(), config.New())
			if err != nil {
				fmt.Printf("    (error getting models: %v)\n", err)
			}
			fmt.Printf("    Models:\n    - %s\n", strings.Join(models, "\n    - "))
		}
		_, _ = fmt.Printf("Selected Inference Provider: %s\n", discoveredFeatures.Inferences[0].Attributes().Name())
		_, _ = fmt.Printf("Available Tools Providers:\n")
		for _, provider := range discoveredFeatures.Tools {
			fmt.Printf("  - %s\n", provider.Attributes().Name())
		}
	}
	return nil
}
