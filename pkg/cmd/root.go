package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/manusa/ai-cli/pkg/ai"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/ui"
	"github.com/manusa/ai-cli/pkg/version"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
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
			if err := o.Run(cmd); err != nil {
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
func (o *AiCliOptions) Run(cmd *cobra.Command) error {

	if o.Version {
		_, _ = fmt.Printf("%s\n", version.Version)
		return nil
	}

	cfg := config.New() // TODO, will need to infer or load from a file

	geminiCli, err := genai.NewClient(cmd.Context(), &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	llm, err := gemini.NewChatModel(cmd.Context(), &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
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
