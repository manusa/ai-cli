// Package ai provides a simple interface to interact with AI models.
// TODO: This is an early stage PoC package
package ai

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/manusa/ai-cli/pkg/api"
	"github.com/mark3labs/mcp-go/client"
)

type Notification struct{}

type Ai struct {
	inferenceProvider api.InferenceProvider
	toolsProviders    []api.ToolsProvider
	tools             *ToolManager
	mcpClients        []*client.Client
	input             chan api.Message
	Output            chan Notification
	session           *Session
	sessionMutex      sync.RWMutex

	llm *DynamicToolCallingChatModel
}

var _ api.Ai = (*Ai)(nil)

func New(inferenceProvider api.InferenceProvider, toolsProviders []api.ToolsProvider) *Ai {
	session := &Session{}
	if inferenceProvider.SystemPrompt() != "" {
		session.systemPrompt = api.NewSystemMessage(inferenceProvider.SystemPrompt())
	}
	return &Ai{
		inferenceProvider: inferenceProvider,
		toolsProviders:    toolsProviders,
		input:             make(chan api.Message),
		Output:            make(chan Notification),
		session:           session,
		sessionMutex:      sync.RWMutex{},
	}
}

func (a *Ai) InferenceAttributes() api.InferenceAttributes {
	return a.inferenceProvider.Attributes()
}

func (a *Ai) ToolCount() int {
	if a.tools == nil {
		return 0
	}
	return a.tools.ToolCount()
}

func (a *Ai) Input() chan api.Message {
	return a.input
}

func (a *Ai) Session() api.Session {
	a.sessionMutex.RLock()
	defer a.sessionMutex.RUnlock()
	sessionShallowCopy := *a.session
	return &sessionShallowCopy
}

// notify sends a notification to the Output channel to inform the UI about changes in the session state.
func (a *Ai) notify() {
	go func() { a.Output <- Notification{} }()
}

func (a *Ai) appendMessage(message api.Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.messages = append(a.session.messages, message)
	a.notify()
}

func (a *Ai) setMessageInProgress(message api.Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.messageInProgress = message
	a.notify()
}

func (a *Ai) setError(err error) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.error = err
	a.notify()
}

func (a *Ai) setRunning(running bool) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.running = running
	a.notify()
}

// Reset resets the AI session, keeping the system prompt intact.
func (a *Ai) Reset() {
	a.sessionMutex.Lock()
	a.tools.EnabledToolsReset()
	defer a.sessionMutex.Unlock()
	a.session = &Session{
		systemPrompt: a.session.SystemPrompt(),
	}
	a.notify()
}

func (a *Ai) Run(ctx context.Context) (err error) {
	// Inference Provider (LLM)
	a.llm, err = NewDynamicToolCallingChatModel(a.inferenceProvider.GetInference(ctx))
	if err != nil {
		return fmt.Errorf("failed to get inference: %w", err)
	}
	// Tools Providers + MCP
	tools := make([]tool.InvokableTool, 0)
	tools = append(tools, toInvokableTools(ctx, a.toolsProviders)...)
	a.mcpClients = startMcpClients(ctx, a.toolsProviders)
	tools = append(tools, mcpClientTools(ctx, a.mcpClients)...)
	a.tools = NewToolManager(ctx, tools)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case userInput := <-a.Input():
				a.prompt(ctx, userInput)
			}
		}
	}()
	return
}

func (a *Ai) Close() {
	stopMcpClients(a.mcpClients)
}

// Prompt sends a prompt to the AI model.
// TODO: Just a PoC
func (a *Ai) prompt(ctx context.Context, userInput api.Message) {
	a.setRunning(true)
	defer func() { a.setRunning(false) }()
	a.setError(nil) // Clear previous error
	a.appendMessage(userInput)
	reActAgent, err := NewReActAgent(ctx, a)
	if err != nil {
		a.setError(err)
		a.setRunning(false)
		return
	}
	// Send PROMPT
	stream, err := reActAgent.Stream(ctx)
	if err != nil {
		a.setError(err)
		a.setRunning(false)
		return
	}
	// Process the stream
	streamedResponse := strings.Builder{}
	defer stream.Close()
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			// End of stream
			break
		}
		if err != nil {
			a.setError(err)
			a.setRunning(false)
			return
		}
		streamedResponse.WriteString(message.Content)
		a.setMessageInProgress(api.NewAssistantMessage(streamedResponse.String())) // Partial message
	}
	a.setRunning(false)
	if streamedResponse.Len() != 0 {
		assistantMessage := api.NewAssistantMessage(streamedResponse.String())
		a.appendMessage(assistantMessage)
	}
	a.setMessageInProgress(api.NewAssistantMessage(""))
}

func (a *Ai) schemaMessages() []*schema.Message {
	session := a.Session()
	var schemaMessages []*schema.Message
	if session.SystemPrompt().Text != "" {
		schemaMessages = append(schemaMessages, schema.SystemMessage(session.SystemPrompt().Text))
	}
	for _, message := range session.Messages() {
		switch message.Type {
		case api.MessageTypeUser:
			schemaMessages = append(schemaMessages, schema.UserMessage(message.Text))
		case api.MessageTypeAssistant:
			schemaMessages = append(schemaMessages, schema.AssistantMessage(message.Text, nil))
		case api.MessageTypeTool:
			randomId := uuid.New().String()
			schemaMessages = append(schemaMessages, schema.AssistantMessage("", []schema.ToolCall{
				{ID: randomId, Type: "function", Function: schema.FunctionCall{Name: message.ToolName, Arguments: "{}"}},
			}))
			schemaMessages = append(schemaMessages, schema.ToolMessage(message.Text, randomId, schema.WithToolName(message.ToolName)))
		}
	}
	return schemaMessages
}
