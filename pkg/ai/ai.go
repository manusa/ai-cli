// Package ai provides a simple interface to interact with AI models.
// TODO: This is an early stage PoC package
package ai

import (
	"context"
	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/manusa/ai-cli/pkg/config"
	"strings"
	"sync"
)

type Notification struct{}

type Ai struct {
	config *config.Config
	llm    gollm.Client
	Input  chan Message
	Output chan Notification
	/*
		sessionChat and session must be kept in sync.
		sessionChat is used internally by gollm and doesn't provide an API to access the chat history
		session is our internal model with the chat history and state
		since both histories must be synchronized, whenever our session is reset, sessionChat must be reset too
	*/
	sessionChat  gollm.Chat
	session      *Session
	sessionMutex sync.RWMutex
}

func New(llm gollm.Client, cfg *config.Config) *Ai {
	session := &Session{
		systemPrompt: NewSystemMessage("You are a helpful AI assistant."),
	}
	return &Ai{
		config:       cfg,
		llm:          llm,
		Input:        make(chan Message),
		Output:       make(chan Notification),
		sessionChat:  llm.StartChat(session.SystemPrompt(), cfg.GeminiModel),
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
	// Send PROMPT
	stream, err := a.sessionChat.SendStreaming(ctx, userInput.Text)
	if err != nil {
		a.setError(err)
		a.setRunning(false)
		return
	}
	// Process the stream
	streamedResponse := strings.Builder{}
	for response, err := range stream {
		if err != nil {
			a.setError(err)
			a.setRunning(false)
			return
		}
		if response == nil {
			// End of stream
			break
		}
		if len(response.Candidates()) == 0 {
			// No candidates, continue to next response ??? TODO
			continue
		}
		for _, part := range response.Candidates()[0].Parts() {
			if text, ok := part.AsText(); ok {
				streamedResponse.WriteString(text)
				a.setMessageInProgress(NewAssistantMessage(streamedResponse.String())) // Partial message
			}
		}
	}
	if streamedResponse.Len() != 0 {
		assistantMessage := NewAssistantMessage(streamedResponse.String())
		a.appendMessage(assistantMessage)
	}
	a.setMessageInProgress(NewAssistantMessage(""))
	a.setRunning(false)
}
