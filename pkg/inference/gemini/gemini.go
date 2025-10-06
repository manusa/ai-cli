package gemini

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/manusa/ai-cli/pkg/config"
	"github.com/manusa/ai-cli/pkg/inference"
	"github.com/manusa/ai-cli/pkg/keyring"
	"github.com/manusa/ai-cli/pkg/ui/components/password_input"
	"google.golang.org/genai"
)

type Provider struct {
	api.BasicInferenceProvider
}

const (
	API_KEY_ENV_VAR = "GEMINI_API_KEY"
)

var (
	defaultModel = "gemini-2.0-flash"
)

var _ api.InferenceProvider = &Provider{}

func (p *Provider) Initialize(ctx context.Context) {
	// TODO: probably move to features.Discover orchestration
	if cfg := config.GetConfig(ctx); cfg != nil {
		p.InferenceParameters = cfg.InferenceParameters(p.Attributes().Name())
	}

	p.Available = p.getApiKey() != ""
	if p.Available {
		p.IsAvailableReason = fmt.Sprintf("%s is set", API_KEY_ENV_VAR)
		p.ProviderModels = []string{defaultModel}
		p.Model = &defaultModel
	} else {
		p.IsAvailableReason = fmt.Sprintf("%s is not set", API_KEY_ENV_VAR)
	}
}

func (p *Provider) GetInference(ctx context.Context) (model.ToolCallingChatModel, error) {
	geminiCli, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: p.getApiKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return gemini.NewChatModel(ctx, &gemini.Config{Client: geminiCli, Model: defaultModel})
}

func (p *Provider) getApiKey() string {
	if key, err := keyring.GetKey(API_KEY_ENV_VAR); err == nil && len(key) > 0 {
		return key
	}
	return os.Getenv(API_KEY_ENV_VAR)
}

func (p *Provider) SystemPrompt() string {
	// Adapted from https://github.com/google-gemini/gemini-cli/blob/5c2bb990d895254e6563acfd26946c389125387f/packages/core/src/core/prompts.ts#L50
	return fmt.Sprintf(`
You are an interactive CLI agent specializing in general tasks.
Your primary goal is to help users safely and efficiently, adhering to the following instructions, company policies, and utilizing your available tools that you can enable at any time.
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
- **Formatting:** Use GitHub-flavored Markdown. Responses will be rendered in monospace. You are able to convert structured output (e.g. JSON, XML, YAML, CSV, etc.) into Markdown or other convenient formats for better readability. Remember to surround code blocks with triple backticks and specify the language when appropriate (e.g., `+"```json"+`).
- **Tools vs. Text:** Use tools for actions, text output *only* for communication. Do not add explanatory comments within tool calls or code blocks unless specifically part of the required code/command itself.
- **Handling Inability:** If unable/unwilling to fulfill a request, state so briefly (1-2 sentences) without excessive justification. Offer alternatives if appropriate.

## Security and Safety Rules
- **Explain Critical Commands:** Before executing commands with '${ShellTool.Name}' that modify the file system, codebase, or system state, you *must* provide a brief explanation of the command's purpose and potential impact. Prioritize user understanding and safety. You should not ask permission to use the tool; the user will be presented with a confirmation dialogue upon use (you do not need to tell them this).
- **Security First:** Always apply security best practices. Never introduce code that exposes, logs, or commits secrets, API keys, or other sensitive information.

## Tool Usage
- **Tool Catalogue:** You are presented with a catalogue of available tools. You can enable any tool you deem useful to fulfill the user's request at any time.
- **Tool Enabling:** Tools need to be enabled, you don't need to ask for permission to enable a tool. Enable the tool you consider most appropriate for the task and continue with the task **No Chitchat**. You can enable tools at any time.
- **Parallelism:** Tools are executed sequentially. You may request multiple tool calls, but they will be executed one at a time in the order you provide.

## URL handling
- **Browsing:** You have access to a web browsing tool. Enable it when user asks to open a URL.
- **URL Extraction:** When the user provides a URL, it might be incomplete, try to infer the complete URL (e.g. prepend the protocol 'https://').

	`, time.Now().Format("January 2, 2006"))
}

func (p *Provider) InstallHelp() error {
	fmt.Printf("To access Gemini, you need to have a Gemini API key.\n")
	fmt.Printf("Get your API key from Google AI Studio (https://aistudio.google.com/api-keys), or from your company.\n")
	fmt.Printf("Paste your API key below:\n")
	apiKey, err := password_input.Prompt()
	if err != nil {
		return err
	}
	return keyring.SetKey(API_KEY_ENV_VAR, apiKey)
}

var instance = &Provider{
	api.BasicInferenceProvider{
		BasicInferenceAttributes: api.BasicInferenceAttributes{
			BasicFeatureAttributes: api.BasicFeatureAttributes{
				FeatureName:        "gemini",
				FeatureDescription: "Google Gemini inference provider",
			},
			LocalAttr:         false,
			PublicAttr:        true,
			SupportsSetupAttr: true,
		},
	},
}

func init() {
	inference.Register(instance)
}
