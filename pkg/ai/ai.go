// Package ai provides a simple interface to interact with AI models.
// TODO: This is an early stage PoC package
package ai

import (
	"context"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	callbackutils "github.com/cloudwego/eino/utils/callbacks"
	"github.com/manusa/ai-cli/pkg/config"
	"io"
	"strings"
	"sync"
)

type Notification struct{}

type Ai struct {
	config       *config.Config
	llm          model.ToolCallingChatModel
	Input        chan Message
	Output       chan Notification
	session      *Session
	sessionMutex sync.RWMutex
}

func New(llm model.ToolCallingChatModel, cfg *config.Config) *Ai {
	session := &Session{}
	return &Ai{
		config:       cfg,
		llm:          llm,
		Input:        make(chan Message),
		Output:       make(chan Notification),
		session:      session,
		sessionMutex: sync.RWMutex{},
	}
}

func (a *Ai) Session() *Session {
	a.sessionMutex.RLock()
	defer a.sessionMutex.RUnlock()
	sessionShallowCopy := *a.session
	return &sessionShallowCopy
}

// notify sends a notification to the Output channel to inform the UI about changes in the session state.
func (a *Ai) notify() {
	go func() { a.Output <- Notification{} }()
}

func (a *Ai) appendMessage(message Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.messages = append(a.session.messages, message)
	a.notify()
}

func (a *Ai) setMessageInProgress(message Message) {
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

func (a *Ai) Run(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case userInput := <-a.Input:
				a.prompt(ctx, userInput)
			}
		}
	}()
	return nil
}

// Prompt sends a prompt to the AI model.
// TODO: Just a PoC
func (a *Ai) prompt(ctx context.Context, userInput Message) {
	a.setRunning(true)
	defer func() { a.setRunning(false) }()
	a.appendMessage(userInput)
	reactAgent, err := a.setUpAgent(ctx)
	if err != nil {
		a.setError(err)
		a.setRunning(false)
		return
	}
	// Send PROMPT
	stream, err := reactAgent.Stream(
		ctx,
		a.schemaMessages(),
		agent.WithComposeOptions(compose.WithCallbacks(callbackutils.NewHandlerHelper().Tool(&callbackutils.ToolCallbackHandler{
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
				a.appendMessage(NewToolMessage(info.Name))
				return ctx
			},
		}).Handler())),
	)
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
		a.setMessageInProgress(NewAssistantMessage(streamedResponse.String())) // Partial message
	}
	a.setRunning(false)
	if streamedResponse.Len() != 0 {
		assistantMessage := NewAssistantMessage(streamedResponse.String())
		a.appendMessage(assistantMessage)
	}
	a.setMessageInProgress(NewAssistantMessage(""))
}

func (a *Ai) schemaMessages() []*schema.Message {
	session := a.Session()
	var schemaMessages []*schema.Message
	if session.SystemPrompt() != "" {
		schemaMessages = append(schemaMessages, schema.SystemMessage(session.SystemPrompt()))
	}
	for _, message := range session.Messages() {
		switch message.Type {
		case MessageTypeUser:
			schemaMessages = append(schemaMessages, schema.UserMessage(message.Text))
		case MessageTypeAssistant:
			schemaMessages = append(schemaMessages, schema.AssistantMessage(message.Text, nil))
		case MessageTypeTool:
			// TODO
			//schemaMessages = append(schemaMessages, schema.ErrorMessage(message.Text))
		}
	}
	return schemaMessages
}

func (a *Ai) setUpAgent(ctx context.Context) (*react.Agent, error) {
	llmWithTools, err := a.llm.WithTools([]*schema.ToolInfo{&FileList.ToolInfo})
	if err != nil {
		return nil, err
	}
	return react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: llmWithTools,
		MaxStep:          10,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{*FileList},
		},
	})
}
