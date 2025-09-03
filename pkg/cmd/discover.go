package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	mcpconfig "github.com/manusa/ai-cli/pkg/mcp-config"
	"github.com/manusa/ai-cli/pkg/mcp-config/cursor"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type DiscoverCmdOptions struct {
	outputFormat string
	mcpConfig    string
	policiesFile string
}

var (
	editors = []string{"cursor"}
)

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
			if err := o.Validate(cmd); err != nil {
				return err
			}
			if err := o.Run(cmd); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", "json", "Output format (json, text)")
	cmd.Flags().StringVar(&o.mcpConfig, "mcp-config", "", fmt.Sprintf("Configure editor MCP config (%s). This option replaces the normal output", strings.Join(editors, ", ")))
	cmd.Flags().StringVar(&o.policiesFile, "policies", "", "Policies file to use")

	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *DiscoverCmdOptions) Complete(cmd *cobra.Command, _ []string) error {
	cfg := config.New()
	cfg.ToolsParameters = tools.GetDefaultParameters()
	cmd.SetContext(config.WithConfig(cmd.Context(), cfg))

	var userPolicies *api.Policies
	if len(o.policiesFile) > 0 {
		var err error
		userPolicies, err = policies.PoliciesProvider.Read(o.policiesFile)
		if err != nil {
			return fmt.Errorf("failed to read preferences: %w", err)
		}
	}
	cmd.SetContext(policies.WithPolicies(cmd.Context(), userPolicies))

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *DiscoverCmdOptions) Validate(cmd *cobra.Command) error {
	if cmd.Flags().Changed("mcp-config") && !slices.Contains(editors, o.mcpConfig) {
		return fmt.Errorf("invalid editor name '%s', must be one of (%s)", o.mcpConfig, strings.Join(editors, ", "))
	}
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *DiscoverCmdOptions) Run(cmd *cobra.Command) error {

	discoveredFeatures := features.Discover(cmd.Context())

	if o.mcpConfig != "" {
		var mcpConfigProvider api.MCPConfig
		if o.mcpConfig == "cursor" {
			mcpConfigProvider = &cursor.CursorMcpConfig{}
		} else {
			return fmt.Errorf("invalid editor name '%s', must be one of (%s)", o.mcpConfig, strings.Join(editors, ", "))
		}
		return mcpconfig.Save(mcpConfigProvider, discoveredFeatures.Tools)
	}

	switch o.outputFormat {
	case "json":
		jsonString, err := discoveredFeatures.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal discovered features to JSON: %w", err)
		}
		_, _ = fmt.Printf("%s\n", jsonString)
	case "text":
		_, _ = fmt.Print(discoveredFeatures.ToHumanReadable())
	}
	return nil
}
