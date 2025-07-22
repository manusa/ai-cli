// Package ai provides a simple interface to interact with AI models.
// TODO: This is an early stage PoC package
package ai

import (
	"context"
	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/manusa/ai-cli/pkg/config"
	"strings"
	"sync"
	"time"
)

type Notification struct{}

type Ai struct {
	ctx    context.Context
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

func New(ctx context.Context, cfg *config.Config) (*Ai, error) {
	llm, err := gollm.NewGeminiAPIClient(ctx, gollm.GeminiAPIClientOptions{
		APIKey: cfg.GoogleApiKey,
	})
	if err != nil {
		return nil, err
	}
	session := &Session{
		systemPrompt: NewSystemMessage("You are a helpful AI assistant."),
	}
	return &Ai{
		ctx:          ctx,
		config:       cfg,
		llm:          llm,
		Input:        make(chan Message),
		Output:       make(chan Notification),
		sessionChat:  llm.StartChat(session.SystemPrompt(), cfg.GeminiModel),
		session:      session,
		sessionMutex: sync.RWMutex{},
	}, nil
}

func (a *Ai) Session() *Session {
	a.sessionMutex.RLock()
	defer a.sessionMutex.RUnlock()
	sessionShallowCopy := *a.session
	return &sessionShallowCopy
}

func (a *Ai) Notify() {
	go func() { a.Output <- Notification{} }()
}

func (a *Ai) appendMessage(message Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.messages = append(a.session.messages, message)
}

func (a *Ai) setMessageInProgress(message Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.messageInProgress = message
	a.Notify()
}

func (a *Ai) setRunning(running bool) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.running = running
	a.Notify()
}

func (a *Ai) Run() error {
	go func() {
		for {
			if !a.Session().IsRunning() {
				select {
				case <-a.ctx.Done():
					return
				case userInput := <-a.Input:
					err := a.prompt(userInput)
					if err != nil {
						// TODO: Figure out a nice way to display errors to the user
						continue
					}
				}
			} else {
				// Loop while running
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	return nil
}

// Prompt sends a prompt to the AI model.
// TODO: Just a PoC
func (a *Ai) prompt(userInput Message) error {
	a.setRunning(true)
	defer func() { a.setRunning(false) }()
	a.appendMessage(userInput)
	// Send PROMPT
	stream, err := a.sessionChat.SendStreaming(a.ctx, userInput.Text)
	if err != nil {
		a.setRunning(false)
		return err
	}
	// Process the stream
	streamedResponse := strings.Builder{}
	for response, err := range stream {
		if err != nil {
			return err
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
	return nil
}
