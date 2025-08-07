package cmd

import (
	"fmt"
	"log"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/features"
	"github.com/manusa/ai-cli/pkg/ui"
	"github.com/spf13/cobra"
)

type ChatCmdOptions struct {
	inference string
	model     string
	tools     []string
	notools   bool
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
	cmd.Flags().StringVar(&o.model, "model", "", "Model to use")
	cmd.Flags().StringSliceVar(&o.tools, "tools", []string{}, "Comma separated list of tools to use, by default all discovered tools will be used")
	err := cmd.Flags().MarkHidden("tools")
	if err != nil {
		log.Fatalln("tools flag is not defined")
	}
	cmd.Flags().BoolVar(&o.notools, "notools", false, "Do not use tools")
	err = cmd.Flags().MarkHidden("notools")
	if err != nil {
		log.Fatalln("notools flag is not defined")
	}
	return cmd
}

// Complete fills in any missing information by gathering data from flags, environment, or other sources
// It converts user input into a usable configuration
func (o *ChatCmdOptions) Complete(_ *cobra.Command, _ []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *ChatCmdOptions) Validate() error {
	return nil
}

// Run executes the main logic of the command once its complete and validated
func (o *ChatCmdOptions) Run(cmd *cobra.Command) error {
	fmt.Printf("tools: %d %v\n", len(o.tools), o.tools)
	cfg := config.New() // TODO, will need to infer or load from a file

	if o.inference != "" {
		cfg.Inference = &o.inference
	}
	if o.model != "" {
		cfg.Model = &o.model
	}

	availableFeatures := features.Discover(cfg)
	if availableFeatures.Inference == nil {
		return fmt.Errorf("no suitable inference found")
	}
	llm, err := (*availableFeatures.Inference).GetInference(cmd.Context(), cfg)
	if err != nil {
		return fmt.Errorf("failed to get inference: %w", err)
	}
	var allTools []*api.Tool
	for _, toolProvider := range availableFeatures.Tools {
		if !useTool(toolProvider.Attributes().Name(), o.notools, o.tools) {
			continue
		}
		tools, err := toolProvider.GetTools(cmd.Context(), cfg)
		if err != nil {
			return fmt.Errorf("failed to get tools from provider %s: %w", toolProvider.Attributes().Name(), err)
		}
		allTools = append(allTools, tools...)
		fmt.Printf("using tool: %s\n", toolProvider.Attributes().Name())
	}
	aiAgent := ai.New(llm, allTools, cfg)
	if err = aiAgent.Run(cmd.Context()); err != nil {
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
	if _, err = p.Run(); err != nil {
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
