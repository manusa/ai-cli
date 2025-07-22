// Package ai provides a simple interface to interact with AI models.
// TODO: This is an early stage PoC package
package ai

import (
	"context"
	"github.com/manusa/ai-cli/pkg/config"
	"sync"
)

type Ai struct {
	config       *config.Config
	Input        chan Message
	Output       chan any
	session      *Session
	sessionMutex sync.RWMutex
}

func New(cfg *config.Config) *Ai {
	return &Ai{
		config:       cfg,
		Input:        make(chan Message),
		Output:       make(chan any, 10),
		session:      &Session{},
		sessionMutex: sync.RWMutex{},
	}
}

func (a *Ai) Session() *Session {
	a.sessionMutex.RLock()
	defer a.sessionMutex.RUnlock()
	sessionShallowCopy := *a.session
	return &sessionShallowCopy
}

func (a *Ai) appendMessage(message Message) {
	a.sessionMutex.Lock()
	defer a.sessionMutex.Unlock()
	a.session.Messages = append(a.session.Messages, message)
}

func (a *Ai) Run(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case userInput := <-a.Input:
				a.appendMessage(userInput)
			}
		}
	}()
	return nil
}

//func (a *Ai) Chat(ctx context.Context) {
//	gollm.NewGeminiAPIClient(ctx, gollm.GeminiAPIClientOptions{
//		APIKey:
//	})
//}
