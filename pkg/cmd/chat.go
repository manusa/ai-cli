package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/ui"
	"github.com/spf13/cobra"
)

type ChatCmdOptions struct{}

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

	cfg := config.New() // TODO, will need to infer or load from a file

	llm, err := inference.Discover(cmd.Context(), cfg)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}
	aiAgent := ai.New(llm, cfg)
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
