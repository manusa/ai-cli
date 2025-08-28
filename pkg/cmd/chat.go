package cmd

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/policies"
	"github.com/manusa/ai-cli/pkg/ui"
	"github.com/spf13/cobra"
)

type ChatCmdOptions struct {
	inference    string
	model        string
	policiesFile string
	tools        []string
	notools      bool

	cfg          *config.Config
	features     *features.Features
	enabledTools []*api.Tool
}

func NewChatCmdOptions() *ChatCmdOptions {
	return &ChatCmdOptions{}
}

func NewChatCmd() *cobra.Command {
	o := NewChatCmdOptions()
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Chat with model",
		Long:  "Start an interactive chat with an AI model",
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

	cmd.Flags().StringVar(&o.inference, "inference", "", "Inference server to use")
	_ = cmd.Flags().MarkHidden("inference") // TODO: evaluate which flags should be exposed
	cmd.Flags().StringVar(&o.model, "model", "", "Model to use")
	_ = cmd.Flags().MarkHidden("model") // TODO: evaluate which flags should be exposed
	cmd.Flags().StringVar(&o.policiesFile, "policies", "", "Policies file to use")
	_ = cmd.Flags().MarkHidden("policies") // TODO: evaluate which flags should be exposed
	cmd.Flags().StringSliceVar(&o.tools, "tools", []string{}, "Comma separated list of tools to use, by default all discovered tools will be used")
	_ = cmd.Flags().MarkHidden("tools")
	cmd.Flags().BoolVar(&o.notools, "notools", false, "Do not use tools")
	_ = cmd.Flags().MarkHidden("notools")
	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *ChatCmdOptions) Complete(cmd *cobra.Command, _ []string) error {
	o.cfg = config.New() // TODO, will need to infer or load from a file

	if o.inference != "" {
		o.cfg.Inference = &o.inference
	}
	if o.model != "" {
		o.cfg.Model = &o.model
	}

	var userPolicies *policies.Policies
	if len(o.policiesFile) > 0 {
		var err error
		userPolicies, err = policies.Read(o.policiesFile)
		if err != nil {
			return fmt.Errorf("failed to read preferences: %w", err)
		}
	}

	o.features = features.Discover(o.cfg, userPolicies)

	for _, toolProvider := range o.features.Tools {
		if !useTool(toolProvider.Attributes().Name(), o.notools, o.tools) {
			continue
		}
		tools, err := toolProvider.GetTools(cmd.Context(), o.cfg)
		if err != nil {
			return fmt.Errorf("failed to get tools from provider %s: %w", toolProvider.Attributes().Name(), err)
		}
		o.enabledTools = append(o.enabledTools, tools...)
	}
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *ChatCmdOptions) Validate() error {
	if o.features.Inference == nil {
		return fmt.Errorf("no suitable inference found")
	}
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *ChatCmdOptions) Run(cmd *cobra.Command) error {
	aiAgent := ai.New(o.cfg, *o.features.Inference, o.enabledTools)
	if err := aiAgent.Run(cmd.Context()); err != nil {
		return fmt.Errorf("failed to run AI: %w", err)
	}
	p := tea.NewProgram(
		ui.NewModel(aiAgent),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithReportFocus(),
	)
	// Agent-UI synchronization
	go func() {
		for {
			select {
			case <-cmd.Context().Done():
				return
			case msg, ok := <-aiAgent.Output:
				if !ok {
					return
				}
				p.Send(msg)
			}
		}
	}()
	// Run TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}
	return nil
}

func useTool(toolName string, notools bool, toolsToUse []string) bool {
	if notools {
		return false
	}
	if len(toolsToUse) == 0 {
		return true
	}
	return slices.Contains(toolsToUse, toolName)
}
