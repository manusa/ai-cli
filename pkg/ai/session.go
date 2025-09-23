package ai

import "github.com/manusa/ai-cli/pkg/api"

type Session struct {
	systemPrompt      api.Message
	messages          []api.Message
	messageInProgress api.Message
	error             error
	running           bool
}

var _ api.Session = (*Session)(nil)

func (s *Session) HasMessages() bool {
	return len(s.messages) > 0 || s.error != nil || (s.IsRunning() && s.messageInProgress.Text != "")
}

func (s *Session) Messages() []api.Message {
	ret := make([]api.Message, len(s.messages))
	copy(ret, s.messages)
	if s.IsRunning() && s.messageInProgress.Text != "" {
		ret = append(ret, s.messageInProgress)
	}
	if s.error != nil {
		ret = append(ret, api.NewErrorMessage(s.error.Error()))
	}
	return ret
}

func (s *Session) SystemPrompt() api.Message {
	return s.systemPrompt
}

func (s *Session) IsRunning() bool {
	return s.running
}
