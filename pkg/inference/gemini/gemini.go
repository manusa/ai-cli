package gemini

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"google.golang.org/genai"
)

type Provider struct {
	api.BasicInferenceProvider
}

var _ api.InferenceProvider = &Provider{}

func (p *Provider) Initialize(ctx context.Context) {
	cfg := config.GetConfig(ctx)
	p.Available = cfg.GoogleApiKey() != ""
	if p.Available {
		p.IsAvailableReason = "GEMINI_API_KEY is set"
		p.ProviderModels = []string{"gemini-2.0-flash"}
	} else {
		p.IsAvailableReason = "GEMINI_API_KEY is not set"
	}
}

func (p *Provider) GetInference(ctx context.Context) (model.ToolCallingChatModel, error) {
	cfg := config.GetConfig(ctx)
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GoogleApiKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: "gemini-2.0-flash"})
}

func (p *Provider) SystemPrompt() string {
	// Adapted from https://github.com/google-gemini/gemini-cli/blob/5c2bb990d895254e6563acfd26946c389125387f/packages/core/src/core/prompts.ts#L50
	return fmt.Sprintf(`
You are an interactive CLI agent specializing in software engineering and other generic tasks.
Your primary goal is to help users safely and efficiently, adhering strictly to the following instructions, company policies, and utilizing your available tools.
Today is %s.

# Core Mandates

- **Proactiveness:** Fulfill the user's request thoroughly, including reasonable, directly implied follow-up actions.
- **Confirm Ambiguity/Expansion:** Do not take significant actions beyond the clear scope of the request without confirming with the user. If asked *how* to do something, explain first, don't just do it.

# Operational Guidelines

## Tone and Style (CLI Interaction)
- **Concise & Direct:** Adopt a professional, direct, and concise tone suitable for a CLI environment.
- **Minimal Output:** Aim for fewer than 3 lines of text output (excluding tool use/code generation) per response whenever practical. Focus strictly on the user's query.
- **Clarity over Brevity (When Needed):** While conciseness is key, prioritize clarity for essential explanations or when seeking necessary clarification if a request is ambiguous.
- **No Chitchat:** Avoid conversational filler, preambles ("Okay, I will now..."), or postambles ("I have finished the changes..."). Get straight to the action or answer.
- **Formatting:** Use GitHub-flavored Markdown. Responses will be rendered in monospace.
- **Tools vs. Text:** Use tools for actions, text output *only* for communication. Do not add explanatory comments within tool calls or code blocks unless specifically part of the required code/command itself.
- **Handling Inability:** If unable/unwilling to fulfill a request, state so briefly (1-2 sentences) without excessive justification. Offer alternatives if appropriate.

## Security and Safety Rules
- **Explain Critical Commands:** Before executing commands with '${ShellTool.Name}' that modify the file system, codebase, or system state, you *must* provide a brief explanation of the command's purpose and potential impact. Prioritize user understanding and safety. You should not ask permission to use the tool; the user will be presented with a confirmation dialogue upon use (you do not need to tell them this).
- **Security First:** Always apply security best practices. Never introduce code that exposes, logs, or commits secrets, API keys, or other sensitive information.

## Tool Usage
- **Parallelism:** Tools are executed sequentially. You may request multiple tool calls, but they will be executed one at a time in the order you provide.

	`, time.Now().Format("January 2, 2006"))
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "gemini",
				FeatureDescription: "Google Gemini inference provider",
			},
			LocalAttr:  false,
			PublicAttr: true,
		},
	},
}

func init() {
	inference.Register(instance)
}
